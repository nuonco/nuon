package imagesync

import (
	"errors"
	"testing"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// fakeLoader counts lookups so a test can assert that a rule settled the
// outcome before the lookup it would otherwise have needed: the callers'
// activity sequences depend on that short-circuiting.
type fakeLoader struct {
	latestBuildID string
	latestErr     error

	installComponentID string
	deployedBuildID    string
	deployedErr        error

	latestCalls   int
	deployedCalls int
}

func (f *fakeLoader) LatestActiveBuildID(_ workflow.Context, _ string) (string, error) {
	f.latestCalls++
	return f.latestBuildID, f.latestErr
}

func (f *fakeLoader) DeployedBuild(_ workflow.Context, _ string) (string, string, error) {
	f.deployedCalls++
	return f.installComponentID, f.deployedBuildID, f.deployedErr
}

func TestDecide(t *testing.T) {
	for _, tc := range []struct {
		name   string
		dep    Dep
		loader fakeLoader

		wantSync               bool
		wantSkip               SkipReason
		wantBuildID            string
		wantInstallComponentID string
		wantLatestCalls        int
		wantDeployedCalls      int
	}{
		{
			name:     "non-image dep is not synced",
			dep:      Dep{ComponentID: "cmp_1", IsImage: false, InAppConfig: true},
			wantSkip: SkipNotAnImage,
		},
		{
			name:     "dep missing from the pinned app config is left alone",
			dep:      Dep{ComponentID: "cmp_1", IsImage: true, InAppConfig: false},
			wantSkip: SkipNotInAppConfig,
		},
		{
			name:            "dep with no active build has nothing to sync",
			dep:             Dep{ComponentID: "cmp_1", IsImage: true, InAppConfig: true},
			loader:          fakeLoader{latestBuildID: ""},
			wantSkip:        SkipNoActiveBuild,
			wantLatestCalls: 1,
		},
		{
			name:              "dep with no install component has nothing to sync against",
			dep:               Dep{ComponentID: "cmp_1", IsImage: true, InAppConfig: true},
			loader:            fakeLoader{latestBuildID: "bld_new", installComponentID: ""},
			wantSkip:          SkipNoInstallComponent,
			wantLatestCalls:   1,
			wantDeployedCalls: 1,
		},
		{
			name: "install already on the latest build",
			dep:  Dep{ComponentID: "cmp_1", IsImage: true, InAppConfig: true},
			loader: fakeLoader{
				latestBuildID:      "bld_current",
				installComponentID: "ic_1",
				deployedBuildID:    "bld_current",
			},
			wantSkip:          SkipAlreadyCurrent,
			wantLatestCalls:   1,
			wantDeployedCalls: 1,
		},
		{
			name: "install on an older build needs a sync",
			dep:  Dep{ComponentID: "cmp_1", IsImage: true, InAppConfig: true},
			loader: fakeLoader{
				latestBuildID:      "bld_new",
				installComponentID: "ic_1",
				deployedBuildID:    "bld_old",
			},
			wantSync:               true,
			wantBuildID:            "bld_new",
			wantInstallComponentID: "ic_1",
			wantLatestCalls:        1,
			wantDeployedCalls:      1,
		},
		{
			name: "install that has never synced the dep needs a sync",
			dep:  Dep{ComponentID: "cmp_1", IsImage: true, InAppConfig: true},
			loader: fakeLoader{
				latestBuildID:      "bld_new",
				installComponentID: "ic_1",
				deployedBuildID:    "",
			},
			wantSync:               true,
			wantBuildID:            "bld_new",
			wantInstallComponentID: "ic_1",
			wantLatestCalls:        1,
			wantDeployedCalls:      1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loader := tc.loader
			got, err := Decide(nil, tc.dep, &loader)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.NeedsSync != tc.wantSync {
				t.Errorf("NeedsSync = %v, want %v", got.NeedsSync, tc.wantSync)
			}
			if got.Skip != tc.wantSkip {
				t.Errorf("Skip = %q, want %q", got.Skip, tc.wantSkip)
			}
			if got.BuildID != tc.wantBuildID {
				t.Errorf("BuildID = %q, want %q", got.BuildID, tc.wantBuildID)
			}
			// Both callers depend on this: the deploy generator puts it on the
			// sync signal, and the action run uses it as the state-gen target.
			if got.InstallComponentID != tc.wantInstallComponentID {
				t.Errorf("InstallComponentID = %q, want %q", got.InstallComponentID, tc.wantInstallComponentID)
			}
			if loader.latestCalls != tc.wantLatestCalls {
				t.Errorf("LatestActiveBuildID called %d times, want %d", loader.latestCalls, tc.wantLatestCalls)
			}
			if loader.deployedCalls != tc.wantDeployedCalls {
				t.Errorf("DeployedBuild called %d times, want %d", loader.deployedCalls, tc.wantDeployedCalls)
			}
		})
	}
}

func TestDecideSurfacesLoaderErrors(t *testing.T) {
	dep := Dep{ComponentID: "cmp_1", IsImage: true, InAppConfig: true}
	wantErr := errors.New("boom")

	if _, err := Decide(nil, dep, &fakeLoader{latestErr: wantErr}); !errors.Is(err, wantErr) {
		t.Errorf("build lookup error = %v, want %v", err, wantErr)
	}

	if _, err := Decide(nil, dep, &fakeLoader{latestBuildID: "bld_new", deployedErr: wantErr}); !errors.Is(err, wantErr) {
		t.Errorf("install component lookup error = %v, want %v", err, wantErr)
	}
}

// A sync deploy row is written before its job runs, so a failed one must not
// answer for the build it never landed — otherwise the staleness check reads
// the install as current and the sync is never retried.
func TestSyncFailed(t *testing.T) {
	for status, want := range map[app.InstallDeployStatus]bool{
		app.InstallDeployStatusError:     true,
		app.InstallDeployStatusCancelled: true,
		app.InstallDeployStatusActive:    false,
		app.InstallDeployStatusSyncing:   false,
		app.InstallDeployStatusPlanning:  false,
		app.InstallDeployStatusNoop:      false,
	} {
		if got := syncFailed(status); got != want {
			t.Errorf("syncFailed(%q) = %v, want %v", status, got, want)
		}
	}
}
