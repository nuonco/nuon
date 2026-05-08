package componentsyncimage

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	workerstate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "component-sync-image"

type Signal struct {
	InstallComponentID string
	DeployID           string
	ComponentID        string
	WorkflowStepID     string
	FlowID             string
	SandboxMode        bool
	Role               string

	runnerJobID string
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
	s.FlowID = flowID
}

var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithAutoRetry = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)

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

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
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
		componentBuild, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, s.ComponentID)
		if err != nil {
			return fmt.Errorf("unable to get component build: %w", err)
		}

		typ := app.InstallDeployTypeSync
		installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID:   install.ID,
			ComponentID: s.ComponentID,
			BuildID:     componentBuild.ID,
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
	l, err = log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("syncing oci artifact")
	if err := s.execSync(ctx, install, installDeploy, s.SandboxMode); err != nil {
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
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)
	if stateGenV2 {
		enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   install.ID,
			OwnerType: "installs",
			QueueName: installshelpers.InstallStateManagerQueueName,
			Signal: &statepartialgenerate.Signal{
				InstallID:       install.ID,
				Targets:         statemanager.TargetsForHint(statemanager.HintDeployCompleted, s.InstallComponentID),
				ForceAll:        true,
				TriggeredByID:   installDeploy.ID,
				TriggeredByType: "install_deploys",
			},
		})
		if err != nil {
			return errors.Wrap(err, "unable to hint state manager")
		} else if _, err := queueclient.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID); err != nil {
			return errors.Wrap(err, "unable to await state generation")
		}

	} else {
		if _, err := workerstate.AwaitGenerateState(ctx, &workerstate.GenerateStateRequest{
			InstallID:       install.ID,
			TriggeredByID:   installDeploy.ID,
			TriggeredByType: "install_deploys",
		}); err != nil {
			return errors.Wrap(err, "unable to generate state")
		}
	}

	return nil
}

func (s *Signal) execSync(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("syncing image into install OCI repository")
	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating sync plan")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		return fmt.Errorf("unable to get build: %w", err)
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	runnerJob, err := activities.AwaitCreateSyncJob(ctx, &activities.CreateSyncJobRequest{
		DeployID:    installDeploy.ID,
		RunnerID:    install.RunnerID,
		Op:          app.RunnerJobOperationTypeExec,
		Type:        build.ComponentConfigConnection.Type.SyncJobType(),
		LogStreamID: logStreamID,
		Metadata: map[string]string{
			"install_id":           install.ID,
			"deploy_id":            installDeploy.ID,
			"install_component_id": installDeploy.InstallComponentID,
			"component_id":         build.ComponentConfigConnection.ComponentID,
			"component_name":       build.ComponentConfigConnection.Component.Name,
		},
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}
	s.runnerJobID = runnerJob.ID

	// create the plan request
	runPlan, err := plan.AwaitCreateSyncPlan(ctx, &plan.CreateSyncPlanRequest{
		InstallID:       install.ID,
		InstallDeployID: installDeploy.ID,
		WorkflowID:      fmt.Sprintf("%s-create-oci-sync-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return errors.Wrap(err, "unable to create plan")
	}

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	// Deprecated: for now we dual write both the plan json and the composite plan
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			SyncOCIPlan: runPlan,
		},
	}); err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	// queue job
	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusSyncing, "executing sync plan")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   install.RunnerID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("%s-execute-job", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to poll job")
		l.Error("error polling sync image job", zap.Error(err))
		return fmt.Errorf("unable to poll job: %w", err)
	}
	l.Info("sync image job was successfully completed")

	// parse outputs
	job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner job")
	}

	var ociArtOutputs state.OCIArtifactOutputs
	if err := mapstructure.Decode(job.ParsedOutputs["image"], &ociArtOutputs); err != nil {
		l.Error("error parsing oci artifact outputs", zap.Error(err))
		return errors.Wrap(err, "unable to parse oci artifact outputs")
	}

	if _, err := activities.AwaitCreateOCIArtifact(ctx, activities.CreateOCIArtifactRequest{
		OwnerID:   installDeploy.ID,
		OwnerType: "install_deploys",
		Outputs:   ociArtOutputs,
	}); err != nil {
		return errors.Wrap(err, "unable to create oci artifact")
	}

	return nil
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
