package componentsyncimage

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/imagesync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/stategen"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "component-sync-image"

type Signal struct {
	InstallComponentID string
	DeployID           string
	ComponentID        string
	// BuildID is the ComponentBuild to sync. When DeployID is empty and
	// BuildID is set, the signal creates an InstallDeploy pinned to this
	// build instead of falling back to "latest build for component" at
	// signal-run time. The workflow generator resolves this at step-gen
	// so the build identity is captured up front.
	BuildID string
	// ComponentConfigConnectionID pins looking up build for specific ccc
	// for that app config id, in case build id is not provided.
	ComponentConfigConnectionID string
	WorkflowStepID              string
	FlowID                      string
	SandboxMode                 bool
	Role                        string

	runnerJobID string
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
	s.FlowID = flowID
}

var (
	_ signal.SignalWithStepContext = (*Signal)(nil)
	_ signal.SignalWithAutoRetry   = (*Signal)(nil)
	_ signal.SignalWithCancel      = (*Signal)(nil)
	_ signal.SignalWithOnRetry     = (*Signal)(nil)
)

func (s *Signal) OnRetry(ctx workflow.Context) error {
	if s.DeployID != "" {
		s.updateDeployStatusWithoutStatusSync(ctx, s.DeployID, app.InstallDeployStatusRetried, "deploy retried")
	}
	return nil
}

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	if s.runnerJobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, s.runnerJobID)
	}
	return nil
}

func (s *Signal) Validate(ctx workflow.Context) error {
	// Validate install component exists
	_, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, s.InstallComponentID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}
	return nil
}

// not every caller pins the connection on the signal, so fall back to the one
// this install's app config resolves to rather than failing the step
func (s *Signal) configConnectionID(ctx workflow.Context, install *app.Install) (string, error) {
	if s.ComponentConfigConnectionID != "" {
		return s.ComponentConfigConnectionID, nil
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return "", errors.Wrap(err, "unable to get app config")
	}

	for _, ccc := range appCfg.ComponentConfigConnections {
		if ccc.ComponentID == s.ComponentID {
			return ccc.ID, nil
		}
	}

	return "", fmt.Errorf("no component config connection for component %s in app config %s", s.ComponentID, install.AppConfigID)
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to create logger")
	}

	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, s.InstallComponentID)
	if err != nil {
		s.updateDeployStatus(ctx, s.DeployID, app.InstallDeployStatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	var installDeploy *app.InstallDeploy
	if s.DeployID != "" {
		installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, s.DeployID)
		if err != nil {
			return errors.Wrap(err, "unable to get deploy")
		}
		s.DeployID = installDeploy.ID
	} else {
		buildID := s.BuildID
		if buildID == "" {
			cccID, err := s.configConnectionID(ctx, install)
			if err != nil {
				return err
			}
			pinned, err := activities.AwaitGetComponentBuildForConfigConnectionByComponentConfigConnectionID(ctx, cccID)
			if err != nil {
				return fmt.Errorf("unable to get pinned component build: %w", err)
			}
			if pinned == nil {
				return fmt.Errorf("no deployable build for component config connection %s", cccID)
			}
			buildID = pinned.ID
		}

		typ := app.InstallDeployTypeSync
		installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID:   install.ID,
			ComponentID: s.ComponentID,
			BuildID:     buildID,
			Type:        typ,
			WorkflowID:  s.FlowID,
			Role:        s.Role,
		})
		if err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}
		s.DeployID = installDeploy.ID
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         s.WorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: "install_deploys",
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		DeployID: s.DeployID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)

	l.Info("syncing oci artifact")
	if err := s.execSync(ctx, install, installDeploy); err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to sync")
		return errors.Wrap(err, "unable to execute sync")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         s.WorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: "install_deploys",
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusActive, "finished")
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	if err := stategen.HintOrGenerate(ctx, stategen.Request{
		StateGenV2:      statemanager.UseStateGenV2(orgEnabled, install.Metadata),
		InstallID:       install.ID,
		Targets:         statemanager.TargetsForHint(statemanager.HintDeployCompleted, s.InstallComponentID),
		ForceAll:        true,
		TriggeredByID:   installDeploy.ID,
		TriggeredByType: "install_deploys",
	}); err != nil {
		return err
	}

	return nil
}

// execSync copies the deploy's image into the install registry. The
// choreography lives in imagesync.RunSyncJob so an image-backed action run,
// which has no workflow step to hang a sync off, applies exactly the same
// steps in the same order.
func (s *Signal) execSync(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy) error {
	return imagesync.RunSyncJob(ctx, imagesync.RunSyncJobRequest{
		Install:       install,
		InstallDeploy: installDeploy,
		Status: func(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string) {
			s.updateDeployStatusWithoutStatusSync(ctx, deployID, status, message)
		},
		OnJobCreated: func(jobID string) {
			s.runnerJobID = jobID
		},
		// Unchanged from when this was inline: these IDs are recorded in the
		// histories of syncs that are still in flight.
		PlanWorkflowID: fmt.Sprintf("%s-create-oci-sync-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
		JobWorkflowID:  fmt.Sprintf("%s-execute-job", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
}

func (s *Signal) updateDeployStatus(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateDeployStatus(ctx, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            status,
		StatusDescription: message,
	}); err != nil {
		l.Error("unable to update deploy status", zap.String("deploy-id", deployID), zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:          deployID,
		Status:            app.Status(status),
		StatusDescription: message,
		SkipStatusSync:    false,
	}); err != nil {
		l.Error("unable to update deploy status v2", zap.String("deploy-id", deployID), zap.Error(err))
	}
}

func (s *Signal) updateDeployStatusWithoutStatusSync(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateDeployStatus(ctx, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            status,
		StatusDescription: message,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update deploy status", zap.String("deploy-id", deployID), zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:          deployID,
		Status:            app.Status(status),
		StatusDescription: message,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update deploy status v2", zap.String("deploy-id", deployID), zap.Error(err))
	}
}
