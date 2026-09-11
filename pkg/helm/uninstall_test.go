package helm

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v4/pkg/action"
	kubefake "helm.sh/helm/v4/pkg/kube/fake"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage"
	"helm.sh/helm/v4/pkg/storage/driver"
)

const confirmReleaseName = "uninstall-me"

// joinedNotFound is the shape helm returns when a real uninstall ran and then
// failed to purge the release record: the sentinel is reachable through an
// Unwrap() []error, not a single Unwrap.
func joinedNotFound() error {
	inner := fmt.Errorf("uninstall: Failed to purge the release: %w", driver.ErrReleaseNotFound)

	return fmt.Errorf("uninstallation completed with 1 error(s): %w", errors.Join(inner))
}

func TestIsReleaseNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"the sentinel itself", driver.ErrReleaseNotFound, true},
		{"wrapped once", fmt.Errorf("get release: %w", driver.ErrReleaseNotFound), true},
		{
			"wrapped several layers deep",
			fmt.Errorf("outer: %w", fmt.Errorf("middle: %w", fmt.Errorf("inner: %w", driver.ErrReleaseNotFound))),
			true,
		},
		{"joined, as a failed purge returns it", joinedNotFound(), true},
		{
			// Some helm paths format the driver error with %s, dropping the chain.
			"message only, chain broken",
			errors.New("uninstall: Release not loaded: acme: release: not found"),
			true,
		},
		{"an unrelated failure", errors.New("etcdserver: request timed out"), false},
		{
			"a not-found that is not about a release",
			errors.New(`secrets "sh.helm.release.v1.acme.v1" not found`),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsReleaseNotFound(tt.err))
		})
	}
}

func confirmConfig(t *testing.T, wrap func(driver.Driver) driver.Driver, revisions ...*release.Release) *action.Configuration {
	t.Helper()

	var d driver.Driver = driver.NewMemory()
	if wrap != nil {
		d = wrap(d)
	}

	store := storage.Init(d)
	for _, rev := range revisions {
		require.NoError(t, store.Create(rev))
	}

	return &action.Configuration{
		Releases:   store,
		KubeClient: &kubefake.PrintingKubeClient{Out: io.Discard, LogOutput: io.Discard},
	}
}

func storedRelease(version int, status release.Status) *release.Release {
	return &release.Release{
		Name:      confirmReleaseName,
		Namespace: "default",
		Version:   version,
		Info:      &release.Info{Status: status},
	}
}

// A store that answers every read but refuses to hand back a release, standing in
// for an unreachable or broken release backend.
type queryFailureDriver struct {
	driver.Driver
}

func (d *queryFailureDriver) Query(_ map[string]string) ([]*release.Release, error) {
	return nil, errors.New("the api server refused the request")
}

func TestConfirmUninstalled(t *testing.T) {
	t.Run("no uninstall error is nothing to resolve", func(t *testing.T) {
		cfg := confirmConfig(t, nil)
		assert.NoError(t, ConfirmUninstalled(cfg, confirmReleaseName, nil))
	})

	t.Run("a real failure is returned untouched", func(t *testing.T) {
		cfg := confirmConfig(t, nil, storedRelease(1, release.StatusDeployed))
		want := errors.New("etcdserver: request timed out")

		assert.Same(t, want, ConfirmUninstalled(cfg, confirmReleaseName, want))
	})

	t.Run("not found and genuinely gone is success", func(t *testing.T) {
		cfg := confirmConfig(t, nil)

		assert.NoError(t, ConfirmUninstalled(cfg, confirmReleaseName, joinedNotFound()))
	})

	t.Run("not found but still stored is a failure", func(t *testing.T) {
		cfg := confirmConfig(t, nil, storedRelease(3, release.StatusUninstalling))
		uninstallErr := joinedNotFound()

		err := ConfirmUninstalled(cfg, confirmReleaseName, uninstallErr)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "still stored at revision 3")
		assert.Contains(t, err.Error(), string(release.StatusUninstalling))
		assert.ErrorIs(t, err, uninstallErr, "the original uninstall error must stay reachable")
	})

	t.Run("not found and unable to re-read is a failure", func(t *testing.T) {
		cfg := confirmConfig(t, func(d driver.Driver) driver.Driver {
			return &queryFailureDriver{Driver: d}
		})

		err := ConfirmUninstalled(cfg, confirmReleaseName, joinedNotFound())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be confirmed removed")
	})
}
