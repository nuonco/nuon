package actionworkflowrun

import (
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

type imageDepInputs struct {
	depIDs         []string
	imageRefNames  map[string]bool
	cccByComponent map[string]*app.ComponentConfigConnection
	loader         *actionDepLoader
}

func (s *Signal) imageDepInputs(ctx workflow.Context, run *app.InstallActionWorkflowRun) (*imageDepInputs, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	if run.ActionWorkflowConfigID.Empty() || run.ActionWorkflowConfig.Image == "" {
		return nil, nil
	}

	depIDs := []string(run.ActionWorkflowConfig.ComponentDependencyIDs)
	if len(depIDs) == 0 {
		return nil, nil
	}

	// ComponentDependencyIDs includes step env/script refs; only sync image
	// components named in the image template.
	imageRefNames := map[string]bool{}
	for _, ref := range refs.ParseFieldRefs(run.ActionWorkflowConfig.Image) {
		if ref.Type == refs.RefTypeComponents {
			imageRefNames[ref.Name] = true
		}
	}
	if len(imageRefNames) == 0 {
		return nil, nil
	}

	if !s.imageActionSupported(run) {
		l.Info("skipping action image dependency sync, image-backed actions are unavailable on this install")
		return nil, nil
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, run.Install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	cccByComponent := make(map[string]*app.ComponentConfigConnection, len(appCfg.ComponentConfigConnections))
	for i := range appCfg.ComponentConfigConnections {
		ccc := &appCfg.ComponentConfigConnections[i]
		cccByComponent[ccc.ComponentID] = ccc
	}

	return &imageDepInputs{
		depIDs:         depIDs,
		imageRefNames:  imageRefNames,
		cccByComponent: cccByComponent,
		loader: &actionDepLoader{
			InstallDeploys: imagesync.InstallDeploys{InstallID: run.InstallID},
			cccByComponent: cccByComponent,
		},
	}, nil
}

func (s *Signal) syncActionImageDeps(
	ctx workflow.Context,
	run *app.InstallActionWorkflowRun,
	logStreamID string,
) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	in, err := s.imageDepInputs(ctx, run)
	if err != nil || in == nil {
		return err
	}
	depIDs, imageRefNames, cccByComponent, loader := in.depIDs, in.imageRefNames, in.cccByComponent, in.loader

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

	if err := stategen.HintOrGenerate(ctx, stategen.Request{
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

func decideImageDeps(ctx workflow.Context, in *imageDepInputs) ([]imageDepCheckpoint, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	var deps []imageDepCheckpoint
	for _, depID := range in.depIDs {
		ccc, inAppConfig := in.cccByComponent[depID]
		if !inAppConfig || ccc == nil || !in.imageRefNames[ccc.Component.Name] {
			continue
		}

		decision, err := imagesync.Decide(ctx, imagesync.Dep{
			ComponentID: depID,
			IsImage:     ccc.Component.Type.IsImage(),
			InAppConfig: true,
		}, in.loader)
		if err != nil {
			return nil, err
		}
		if !decision.NeedsSync {
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

		deps = append(deps, imageDepCheckpoint{
			ComponentID:        depID,
			BuildID:            decision.BuildID,
			InstallComponentID: decision.InstallComponentID,
		})
	}
	return deps, nil
}

func publishImageDeps(ctx workflow.Context, run *app.InstallActionWorkflowRun, deps []imageDepCheckpoint) error {
	if len(deps) == 0 {
		return nil
	}

	syncedTargets := make([]statemanager.PartialTarget, 0, len(deps))
	for _, dep := range deps {
		syncedTargets = append(syncedTargets,
			statemanager.TargetsForHint(statemanager.HintDeployCompleted, dep.InstallComponentID)...)
	}
	var prep imagesync.SyncPreparation
	if _, err := job.LoadCheckpoint(ctx, deps[len(deps)-1].phaseKey()+"-preparation", &prep); err != nil {
		return err
	}

	if err := stategen.HintOrGenerate(ctx, stategen.Request{
		InstallID:       run.InstallID,
		Targets:         syncedTargets,
		ForceAll:        true,
		TriggeredByID:   prep.DeployID,
		TriggeredByType: "install_deploys",
	}); err != nil {
		return errors.Wrap(err, "unable to generate state after image dependency sync")
	}
	return nil
}

func (s *Signal) imageActionSupported(run *app.InstallActionWorkflowRun) bool {
	return supportedImageActionPlatform(run.Install.RunnerGroup.Platform)
}

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
