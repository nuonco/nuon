package imagesync

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

type SkipReason string

const (
	SkipNotAnImage         SkipReason = "not an image component"
	SkipNotInAppConfig     SkipReason = "not in the install's app config"
	SkipNoActiveBuild      SkipReason = "no active build for the pinned app config version"
	SkipNoInstallComponent SkipReason = "install has no component record yet"
	SkipAlreadyCurrent     SkipReason = "install already synced the latest build"
)

type Dep struct {
	ComponentID string
	IsImage     bool
	InAppConfig bool
}

type Loader interface {
	LatestActiveBuildID(ctx workflow.Context, componentID string) (string, error)
	DeployedBuild(ctx workflow.Context, componentID string) (installComponentID string, buildID string, err error)
}

type InstallDeploys struct {
	InstallID string
}

// Failed/cancelled syncs must not count as current: the deploy row is written
// before the job runs, so treating them as landed would skip retries forever.
func (d InstallDeploys) DeployedBuild(ctx workflow.Context, componentID string) (string, string, error) {
	installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   d.InstallID,
		ComponentID: componentID,
	})
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to get install component for image dep %s", componentID)
	}
	if installComp == nil {
		// No install component record yet for this dep — there is no
		// runner-side state to sync against. The normal install
		// bootstrapping flow is responsible for creating it; skip
		// silently here.
		return "", "", nil
	}
	if len(installComp.InstallDeploys) == 0 {
		return installComp.ID, "", nil
	}

	// AwaitGetInstallComponent preloads the most recent InstallDeploy
	// (any type, ORDER BY created_at DESC LIMIT 1). For image
	// components every install_deploy is a sync-image, so the most
	// recent deploy is the currently-synced build. When the deployed
	// build matches the app-config-version-pinned latest Active
	// build, no sync is needed.
	latest := installComp.InstallDeploys[0]
	if syncFailed(latest.Status) {
		return installComp.ID, "", nil
	}
	return installComp.ID, latest.ComponentBuildID, nil
}

func syncFailed(status app.InstallDeployStatus) bool {
	switch status {
	case app.InstallDeployStatusError, app.InstallDeployStatusCancelled:
		return true
	default:
		return false
	}
}

type Decision struct {
	NeedsSync          bool
	Skip               SkipReason
	BuildID            string
	InstallComponentID string
}

func (d Decision) WorthLogging() bool {
	return d.Skip == SkipNoActiveBuild || d.Skip == SkipNoInstallComponent
}

func Decide(ctx workflow.Context, dep Dep, loader Loader) (Decision, error) {
	if !dep.IsImage {
		return Decision{Skip: SkipNotAnImage}, nil
	}
	if !dep.InAppConfig {
		return Decision{Skip: SkipNotInAppConfig}, nil
	}

	buildID, err := loader.LatestActiveBuildID(ctx, dep.ComponentID)
	if err != nil {
		return Decision{}, err
	}
	if buildID == "" {
		return Decision{Skip: SkipNoActiveBuild}, nil
	}

	installComponentID, deployedBuildID, err := loader.DeployedBuild(ctx, dep.ComponentID)
	if err != nil {
		return Decision{}, err
	}
	if installComponentID == "" {
		return Decision{Skip: SkipNoInstallComponent}, nil
	}
	if deployedBuildID == buildID {
		return Decision{Skip: SkipAlreadyCurrent}, nil
	}

	return Decision{
		NeedsSync:          true,
		BuildID:            buildID,
		InstallComponentID: installComponentID,
	}, nil
}
