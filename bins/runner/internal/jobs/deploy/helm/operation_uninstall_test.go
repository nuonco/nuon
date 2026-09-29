package helm

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"helm.sh/helm/v4/pkg/action"
	kubefake "helm.sh/helm/v4/pkg/kube/fake"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage"
	"helm.sh/helm/v4/pkg/storage/driver"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func helmHandler(t *testing.T, wrap func(driver.Driver) driver.Driver, revisions ...*release.Release) (*handler, *action.Configuration) {
	t.Helper()

	var d driver.Driver = driver.NewMemory()
	if wrap != nil {
		d = wrap(d)
	}

	store := storage.Init(d)
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

type vanishedOnPurgeDriver struct {
	driver.Driver
}

func (d *vanishedOnPurgeDriver) Delete(key string) (*release.Release, error) {
	if _, err := d.Driver.Delete(key); err != nil {
		return nil, err
	}

	return nil, driver.ErrReleaseNotFound
}

type phantomNotFoundDriver struct {
	driver.Driver
}

func (d *phantomNotFoundDriver) Delete(_ string) (*release.Release, error) {
	return nil, driver.ErrReleaseNotFound
}

type purgeFailureDriver struct {
	driver.Driver
}

func (d *purgeFailureDriver) Delete(_ string) (*release.Release, error) {
	return nil, errors.New("etcdserver: request timed out")
}

type vanishesAfterFirstReadDriver struct {
	driver.Driver

	reads int
}

func (d *vanishesAfterFirstReadDriver) Query(labels map[string]string) ([]*release.Release, error) {
	d.reads++
	if d.reads > 1 {
		return nil, driver.ErrReleaseNotFound
	}

	return d.Driver.Query(labels)
}

type readFailureDriver struct {
	driver.Driver
}

func (d *readFailureDriver) Query(_ map[string]string) ([]*release.Release, error) {
	return nil, errors.New("the api server refused the request")
}

func observedLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.WarnLevel)

	return zap.New(core), logs
}

func TestUninstall_NoPreviousRelease(t *testing.T) {
	h, cfg := helmHandler(t, nil)

	require.NoError(t, h.uninstall(context.Background(), zap.NewNop(), cfg))
}

func TestUninstall_ReadFailurePropagates(t *testing.T) {
	h, cfg := helmHandler(t, func(d driver.Driver) driver.Driver {
		return &readFailureDriver{Driver: d}
	}, testRevision(1, release.StatusDeployed))

	err := h.uninstall(context.Background(), zap.NewNop(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "before uninstalling it")
}

func TestUninstall_ReleaseRemovedAfterTheReadIsNotAnError(t *testing.T) {
	h, cfg := helmHandler(t, func(d driver.Driver) driver.Driver {
		return &vanishesAfterFirstReadDriver{Driver: d}
	}, testRevision(1, release.StatusDeployed))

	l, logs := observedLogger()
	require.NoError(t, h.uninstall(context.Background(), l, cfg))
	assert.Empty(t, logs.All(), "helm absorbs this one, so there is nothing to warn about")
}

func TestUninstall_PurgeNotFoundIsTreatedAsUninstalled(t *testing.T) {
	h, cfg := helmHandler(t, func(d driver.Driver) driver.Driver {
		return &vanishedOnPurgeDriver{Driver: d}
	}, testRevision(1, release.StatusDeployed))

	l, logs := observedLogger()
	require.NoError(t, h.uninstall(context.Background(), l, cfg))

	warnings := logs.FilterMessageSnippet("no longer stored").All()
	require.Len(t, warnings, 1, "the tolerated not-found must be recorded")
	assert.Equal(t, testReleaseName, warnings[0].ContextMap()["release"])
}

func TestUninstall_NotFoundWithReleaseStillStoredFails(t *testing.T) {
	h, cfg := helmHandler(t, func(d driver.Driver) driver.Driver {
		return &phantomNotFoundDriver{Driver: d}
	}, testRevision(1, release.StatusDeployed))

	err := h.uninstall(context.Background(), zap.NewNop(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "still stored")
}

func TestUninstall_RealFailurePropagates(t *testing.T) {
	h, cfg := helmHandler(t, func(d driver.Driver) driver.Driver {
		return &purgeFailureDriver{Driver: d}
	}, testRevision(1, release.StatusDeployed))

	err := h.uninstall(context.Background(), zap.NewNop(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "etcdserver: request timed out")
}
