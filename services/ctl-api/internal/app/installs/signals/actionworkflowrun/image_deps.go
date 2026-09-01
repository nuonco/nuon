package actionworkflowrun

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/imagesync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/stategen"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// syncActionImageDeps brings the image components an image-backed action's
// image template points at up to date in the install's registry, before the
// run is planned.
//
// The action's image is rendered from install state at plan time. An image dep
// the install is behind on therefore resolves to the digest the install synced
// last, and the action silently runs an older image; a dep the install has
// never synced does not render at all. A deploy already prepends a sync step
// for the image deps of the component it is about to deploy — this is the same
// rule (imagesync.Decide) applied where an action can use it, since an action's
// triggers include ones that generate no workflow steps to prepend to (cron,
// adhoc) and ones whose parent component does not list the action's image dep.
//
// The dep itself is derived at config sync: a component referenced from the
// image template lands in the action config's ComponentDependencyIDs, so an app
// author does not have to declare it.
func (s *Signal) syncActionImageDeps(
	ctx workflow.Context,
	run *app.InstallActionWorkflowRun,
	logStreamID string,
	metadata pgtype.Hstore,
) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	// Only a config-backed action can carry an image — adhoc runs have no
	// action config to carry one.
	if run.ActionWorkflowConfigID.Empty() || run.ActionWorkflowConfig.Image == "" {
		return nil
	}

	depIDs := []string(run.ActionWorkflowConfig.ComponentDependencyIDs)
	if len(depIDs) == 0 {
		return nil
	}

	// Only the components the image template itself references are in scope.
	// ComponentDependencyIDs is the union of every component the action
	// references anywhere — step env vars and inline scripts included — and
	// syncing all of those would push image components into the install that
	// nobody asked to deploy.
	imageRefNames := map[string]bool{}
	for _, ref := range refs.ParseFieldRefs(run.ActionWorkflowConfig.Image) {
		if ref.Type == refs.RefTypeComponents {
			imageRefNames[ref.Name] = true
		}
	}
	if len(imageRefNames) == 0 {
		// A public ref is mirrored per run and needs no component sync, so
		// there is nothing to look up.
		return nil
	}

	// Nothing should be synced for a run that cannot execute on this install
	// anyway. checkImageActionSupported reports the same condition after the
	// plan is built and stays the single source of that error message.
	supported, err := s.imageActionSupported(ctx, run)
	if err != nil {
		return err
	}
	if !supported {
		l.Info("skipping action image dependency sync, image-backed actions are unavailable on this install")
		return nil
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, run.Install.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	cccByComponent := make(map[string]*app.ComponentConfigConnection, len(appCfg.ComponentConfigConnections))
	for i := range appCfg.ComponentConfigConnections {
		ccc := &appCfg.ComponentConfigConnections[i]
		cccByComponent[ccc.ComponentID] = ccc
	}

	loader := &actionDepLoader{
		InstallDeploys: imagesync.InstallDeploys{InstallID: run.InstallID},
		cccByComponent: cccByComponent,
	}

	syncedTargets := make([]statemanager.PartialTarget, 0)
	lastDeployID := ""

	for _, depID := range depIDs {
		ccc, inAppConfig := cccByComponent[depID]
		inAppConfig = inAppConfig && ccc != nil

		if !inAppConfig || !imageRefNames[ccc.Component.Name] {
			continue
		}

		decision, err := imagesync.Decide(ctx, imagesync.Dep{
			ComponentID: depID,
			IsImage:     ccc.Component.Type.IsImage(),
			InAppConfig: true,
		}, loader)
		if err != nil {
			return err
		}
		if !decision.NeedsSync {
			// A skip that means something upstream has not happened is the
			// reason an action would run an older image, so it is reported
			// rather than dropped.
			if decision.WorthLogging() {
				l.Info("action image dependency cannot be synced",
					zap.String("component_id", depID),
					zap.String("component_name", ccc.Component.Name),
					zap.String("reason", string(decision.Skip)))
				continue
			}
			l.Debug("action image dependency needs no sync",
				zap.String("component_id", depID),
				zap.String("reason", string(decision.Skip)))
			continue
		}

		l.Info("syncing image dependency before action run",
			zap.String("component_id", depID),
			zap.String("component_name", ccc.Component.Name),
			zap.String("build_id", decision.BuildID))

		installDeploy, err := imagesync.Sync(ctx, imagesync.SyncRequest{
			Install:           &run.Install,
			ComponentID:       depID,
			BuildID:           decision.BuildID,
			FlowID:            s.InstallWorkflowID,
			ParentLogStreamID: logStreamID,
			OnJobCreated: func(jobID string) {
				s.runnerJobID = jobID
			},
			// One workflow execution can drive several runs and each run
			// several deps, so the child workflows a sync starts need an ID
			// per run and component.
			WorkflowIDSuffix: "-action-image-dep-" + run.ID + "-" + depID,
		})
		if err != nil {
			return errors.Wrapf(err, "unable to sync image dependency %s", depID)
		}

		syncedTargets = append(syncedTargets,
			statemanager.TargetsForHint(statemanager.HintDeployCompleted, decision.InstallComponentID)...)
		lastDeployID = installDeploy.ID
	}

	if len(syncedTargets) == 0 {
		return nil
	}

	// The ref the action's image template renders comes out of install state,
	// so every sync has to be visible there before the plan is built. One
	// regeneration covers them all: only the state as of the last sync matters,
	// and each round trip blocks on a callback that can time out.
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	if err := stategen.HintOrGenerate(ctx, stategen.Request{
		StateGenV2:      statemanager.UseStateGenV2(orgEnabled, metadata),
		InstallID:       run.InstallID,
		Targets:         syncedTargets,
		ForceAll:        true,
		TriggeredByID:   lastDeployID,
		TriggeredByType: "install_deploys",
	}); err != nil {
		return errors.Wrap(err, "unable to generate state after image dependency sync")
	}

	return nil
}

// imageActionSupported reports whether this install can run image-backed
// actions at all.
//
// It keeps "the gate says no" separate from "the gate could not be evaluated":
// treating a failed feature lookup as an unsupported install would skip the
// image sync and let the run go ahead against a stale image.
func (s *Signal) imageActionSupported(ctx workflow.Context, run *app.InstallActionWorkflowRun) (bool, error) {
	enabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureImageBackedActions))
	if err != nil {
		return false, errors.Wrap(err, "unable to check image-backed-actions feature")
	}
	if !enabled {
		return false, nil
	}

	return supportedImageActionPlatform(run.Install.RunnerGroup.Platform), nil
}

// actionDepLoader answers imagesync.Decide's build lookup against the install's
// pinned app config snapshot. The install-side half comes from
// imagesync.InstallDeploys so it cannot drift from the deploy generator's.
type actionDepLoader struct {
	imagesync.InstallDeploys

	cccByComponent map[string]*app.ComponentConfigConnection
}

func (l *actionDepLoader) LatestActiveBuildID(ctx workflow.Context, componentID string) (string, error) {
	ccc, ok := l.cccByComponent[componentID]
	if !ok || ccc == nil {
		return "", nil
	}

	build, err := activities.AwaitGetComponentBuildForConfigConnectionByComponentConfigConnectionID(ctx, ccc.ID)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get pinned build for image dep %s", componentID)
	}
	if build == nil {
		return "", nil
	}
	return build.ID, nil
}
