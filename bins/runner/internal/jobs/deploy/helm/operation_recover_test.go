package helm

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	kubefake "helm.sh/helm/v4/pkg/kube/fake"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage"
	"helm.sh/helm/v4/pkg/storage/driver"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

const testReleaseName = "recover-me"

// recoverHandler builds a handler wired to an in-memory release store and a fake
// cluster, so the decision table can be exercised without a real cluster.
func recoverHandler(t *testing.T, revisions ...*release.Release) (*handler, *action.Configuration) {
	t.Helper()

	store := storage.Init(driver.NewMemory())
	for _, rev := range revisions {
		require.NoError(t, store.Create(rev))
	}

	actionCfg := &action.Configuration{
		Releases:   store,
		KubeClient: &kubefake.PrintingKubeClient{Out: io.Discard, LogOutput: io.Discard},
	}

	h := &handler{
		state: &handlerState{
			plan: &plantypes.DeployPlan{
				HelmDeployPlan: &plantypes.HelmDeployPlan{
					Name:           testReleaseName,
					Namespace:      "default",
					RecoverRelease: true,
				},
			},
			timeout: time.Minute,
		},
	}

	return h, actionCfg
}

func testRevision(version int, status release.Status) *release.Release {
	return &release.Release{
		Name:      testReleaseName,
		Namespace: "default",
		Version:   version,
		Info:      &release.Info{Status: status},
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{Name: "app", Version: "0.1.0", APIVersion: chart.APIVersionV2},
		},
		Config: map[string]interface{}{},
	}
}

// The invariant that makes this safe to expose as a button: a healthy release is
// never touched, no matter how many times the action is run.
func TestRecoverRelease_LeavesHealthyReleaseAlone(t *testing.T) {
	for _, status := range []release.Status{
		release.StatusDeployed,
		release.StatusFailed,
		release.StatusSuperseded,
	} {
		t.Run(string(status), func(t *testing.T) {
			h, cfg := recoverHandler(t, testRevision(1, status))

			res, err := h.recoverRelease(context.Background(), zap.NewNop(), cfg)
			require.NoError(t, err)
			assert.Equal(t, recoverActionNone, res.Action)
			assert.Equal(t, string(status), res.Before)

			history, err := cfg.Releases.History(testReleaseName)
			require.NoError(t, err)
			assert.Len(t, history, 1, "no revision should have been added")
			assert.Equal(t, status, history[0].Info.Status, "status must be untouched")
		})
	}
}

func TestRecoverRelease_NoReleaseStored(t *testing.T) {
	h, cfg := recoverHandler(t)

	res, err := h.recoverRelease(context.Background(), zap.NewNop(), cfg)
	require.NoError(t, err)
	assert.Equal(t, recoverActionNone, res.Action)
	assert.Empty(t, res.Before)
	assert.Contains(t, res.Summary(testReleaseName), "nothing to recover")
}

func TestRecoverRelease_RollsBackToLastGoodRevision(t *testing.T) {
	h, cfg := recoverHandler(t,
		testRevision(1, release.StatusSuperseded),
		testRevision(2, release.StatusDeployed),
		testRevision(3, release.StatusPendingUpgrade),
	)

	res, err := h.recoverRelease(context.Background(), zap.NewNop(), cfg)
	require.NoError(t, err)
	assert.Equal(t, recoverActionRollback, res.Action)
	assert.Equal(t, 2, res.Revision)
	assert.Equal(t, string(release.StatusPendingUpgrade), res.Before)

	last, err := cfg.Releases.Last(testReleaseName)
	require.NoError(t, err)
	assert.Equal(t, release.StatusDeployed, last.Info.Status,
		"the release should be usable again after recovery")
	assert.Equal(t, 4, last.Version, "a rollback records a new revision")
	assert.Contains(t, res.Summary(testReleaseName), "revision 2")
}

// The most common stuck state: a first install that never rolled out. There is no
// revision behind it, so rollback cannot help and the release has to go.
func TestRecoverRelease_UninstallsWhenNoRevisionEverRolledOut(t *testing.T) {
	h, cfg := recoverHandler(t, testRevision(1, release.StatusPendingInstall))

	res, err := h.recoverRelease(context.Background(), zap.NewNop(), cfg)
	require.NoError(t, err)
	assert.Equal(t, recoverActionUninstall, res.Action)
	assert.Equal(t, string(release.StatusPendingInstall), res.Before)
	assert.Contains(t, res.Summary(testReleaseName), "no revision to roll back to")

	last, err := cfg.Releases.Last(testReleaseName)
	if err == nil {
		assert.Equal(t, release.StatusUninstalled, last.Info.Status)
	}
}

// A failed revision is not a safe rollback target, so a release stuck above only
// failed revisions is removed rather than returned to a broken state.
func TestRecoverRelease_FailedRevisionsAreNotRollbackTargets(t *testing.T) {
	h, cfg := recoverHandler(t,
		testRevision(1, release.StatusFailed),
		testRevision(2, release.StatusPendingUpgrade),
	)

	res, err := h.recoverRelease(context.Background(), zap.NewNop(), cfg)
	require.NoError(t, err)
	assert.Equal(t, recoverActionUninstall, res.Action)
}

func TestRecoverRelease_IsIdempotent(t *testing.T) {
	h, cfg := recoverHandler(t,
		testRevision(1, release.StatusDeployed),
		testRevision(2, release.StatusPendingUpgrade),
	)

	first, err := h.recoverRelease(context.Background(), zap.NewNop(), cfg)
	require.NoError(t, err)
	require.Equal(t, recoverActionRollback, first.Action)

	second, err := h.recoverRelease(context.Background(), zap.NewNop(), cfg)
	require.NoError(t, err)
	assert.Equal(t, recoverActionNone, second.Action,
		"a second run must not roll the recovered release back again")
}

// A recovery never fetches the chart archive, so anything Exec touches before it
// branches to the recovery must tolerate a nil archive. Reading the base path off
// it directly panicked on every recovery.
func TestBasePath_ToleratesMissingArchive(t *testing.T) {
	h, _ := recoverHandler(t)
	assert.Empty(t, h.basePath(), "a recovery has no unpacked chart")
	assert.Empty(t, (&handler{}).basePath(), "no state means no base path")
}

func TestIsRecovery(t *testing.T) {
	h, _ := recoverHandler(t)
	assert.True(t, h.isRecovery())

	// The field must be a plain bool: as a struct it becomes a $ref, and a
	// documented $ref field is generated as an inline struct VALUE that decodes
	// non-nil on every deploy, skipping the chart for all of them.
	h.state.plan.HelmDeployPlan.RecoverRelease = false
	assert.False(t, h.isRecovery(), "a normal deploy must not skip the chart")

	assert.False(t, (&handler{}).isRecovery(), "no state means no recovery")
}
