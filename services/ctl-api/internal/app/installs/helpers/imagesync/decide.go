// Package imagesync holds the rules for deciding whether an image component a
// consumer depends on needs to be synced into an install, and the runner-job
// choreography for doing it.
//
// Two callers need the same answer. A deploy prepends a sync step for the image
// deps of the component it is about to deploy. An image-backed action run has
// no step to prepend to — its image is rendered from install state at plan time,
// so a dep that has not been synced resolves to whatever the install synced
// last — and syncs inline before it plans. Keeping the rules in one place is
// what stops the two from drifting apart.
package imagesync

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

// SkipReason names why a dependency does not need a sync. SkipNoActiveBuild and
// SkipNoInstallComponent usually mean something upstream has not happened yet
// rather than that the dep is current, so callers log them rather than dropping
// them silently.
type SkipReason string

const (
	SkipNotAnImage         SkipReason = "not an image component"
	SkipNotInAppConfig     SkipReason = "not in the install's app config"
	SkipNoActiveBuild      SkipReason = "no active build for the pinned app config version"
	SkipNoInstallComponent SkipReason = "install has no component record yet"
	SkipAlreadyCurrent     SkipReason = "install already synced the latest build"
)

// Dep is what the rules need about one dependency that the caller already has
// in memory, without a lookup.
type Dep struct {
	ComponentID string

	// IsImage is false for anything that isn't a docker_build or
	// external_image component: only image components sync into the install
	// registry.
	IsImage bool

	// InAppConfig is false when the dep has no ComponentConfigConnection in
	// the install's pinned app config snapshot, i.e. this app config version
	// does not include the dep at all.
	InAppConfig bool
}

// Loader supplies the two facts that require a lookup. Callers differ on the
// first — the deploy generator reads builds off the app config snapshot cached
// on its generation context, an action run reads them fresh — and share the
// second through InstallDeploys. Decide calls them in order and stops at the
// first one that settles the outcome, so a caller's activity sequence is
// exactly the sequence it had when these rules were inline.
type Loader interface {
	// LatestActiveBuildID returns the latest Active ComponentBuild for the
	// component, pinned to the install's app config version. An empty ID
	// means the component has no deployable build for that version.
	//
	// The lookup must stay pinned to the install's app config version. Each
	// ComponentConfigConnection (ccc) is the per-app-config-version snapshot of
	// a component's config and is what owns that component's builds: when an
	// app config is re-synced, fresh ccc rows are created for every component,
	// and builds that happen later are tied to the new ccc, not the old.
	// Asking for "latest active build for component X" across every ccc
	// over-syncs into installs pinned to an older app config, and decouples the
	// dep from the consumer's snapshot.
	LatestActiveBuildID(ctx workflow.Context, componentID string) (string, error)

	// DeployedBuild returns the install's component record and the build it
	// last synced. An empty install component ID means the install has no
	// record for this component yet, so there is no runner-side state to sync
	// against.
	DeployedBuild(ctx workflow.Context, componentID string) (installComponentID string, buildID string, err error)
}

// InstallDeploys answers the DeployedBuild half of Loader for one install.
// Callers embed it so that half of the rule cannot drift between them.
type InstallDeploys struct {
	InstallID string
}

// DeployedBuild reports the build the install last synced for a component.
//
// A sync deploy row is created before its runner job runs, so the most recent
// deploy names a build that may never have landed. A failed one therefore
// reports "nothing synced": otherwise the failed row would answer for the build
// forever, the staleness check would read the install as current, and the sync
// would never be retried. An in-flight sync is left alone — reporting it as
// missing would start a second sync for the same build.
func (d InstallDeploys) DeployedBuild(ctx workflow.Context, componentID string) (string, string, error) {
	// AwaitGetInstallComponent preloads the most recent InstallDeploy (any
	// type, ORDER BY created_at DESC LIMIT 1). Every install_deploy for an
	// image component is a sync, so the most recent one names the build
	// currently in the install registry.
	//
	// A nil install component means the install has no runner-side state for
	// the dep yet. Install bootstrapping owns creating it, so this reports
	// "none" and lets Decide skip.
	installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   d.InstallID,
		ComponentID: componentID,
	})
	if err != nil {
		return "", "", errors.Wrapf(err, "unable to get install component for image dep %s", componentID)
	}
	if installComp == nil {
		return "", "", nil
	}
	if len(installComp.InstallDeploys) == 0 {
		return installComp.ID, "", nil
	}

	latest := installComp.InstallDeploys[0]
	if syncFailed(latest.Status) {
		return installComp.ID, "", nil
	}
	return installComp.ID, latest.ComponentBuildID, nil
}

// syncFailed reports whether a deploy ended without putting its build in the
// install registry.
func syncFailed(status app.InstallDeployStatus) bool {
	switch status {
	case app.InstallDeployStatusError, app.InstallDeployStatusCancelled:
		return true
	default:
		return false
	}
}

// Decision is the outcome for one dependency. When NeedsSync is set, BuildID
// and InstallComponentID are the ones to sync; otherwise Skip says why not.
type Decision struct {
	NeedsSync          bool
	Skip               SkipReason
	BuildID            string
	InstallComponentID string
}

// Worth logging reports whether a skip means something upstream has not
// happened yet, rather than the install simply being current. Callers surface
// those, because they are the reasons a consumer silently runs an older image.
func (d Decision) WorthLogging() bool {
	return d.Skip == SkipNoActiveBuild || d.Skip == SkipNoInstallComponent
}

// Decide reports whether one image dependency needs syncing into the install.
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
