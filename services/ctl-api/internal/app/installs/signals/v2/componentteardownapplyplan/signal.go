package componentteardownapplyplan

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	installstate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const SignalType signal.SignalType = "component-teardown-apply-plan"

type Signal struct {
	InstallComponentID    string
	ComponentID           string
	FlowID                string
	FlowStepID            string
	InstallWorkflowStepID string
	SandboxMode           bool
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
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
	install, err := activities.AwaitGetInstallForInstallComponentByInstallComponentID(ctx, s.InstallComponentID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	if s.ComponentID == "" {
		return fmt.Errorf("component ID is required")
	}

	installDeploy, err := activities.AwaitGetInstallDeployForApplyStep(ctx, activities.GetInstallDeployForApplyStep{
		InstallWorkflowID: s.FlowID,
		ComponentID:       s.ComponentID,
	})
	if err != nil {
		s.updateDeployStatus(ctx, "", app.InstallDeployStatusError, "unable to get install deploy from previous step")
		return errors.Wrap(err, "unable to get install deploy")
	}

	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			s.updateDeployStatus(updateCtx, installDeploy.ID, app.InstallDeployStatusCancelled, "teardown cancelled")
		}
	}()

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         s.InstallWorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: "install_deploys",
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	ctx = cctx.SetLogStreamWorkflowContext(ctx, &installDeploy.LogStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, installDeploy.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install deploy")
	}

	l.Info("executing plan")
	if err := s.execApplyPlan(ctx, install, installDeploy, s.InstallWorkflowStepID, s.SandboxMode); err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to deploy")
		return errors.Wrap(err, "unable to execute deploy")
	}

	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusInactive, "successfully torn down")
	s.updateInstallComponentStatus(ctx, installDeploy.InstallComponentID, app.InstallComponentStatusInactive, "successfully torn down")

	_, err = installstate.AwaitGenerateState(ctx, &installstate.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   installDeploy.ID,
		TriggeredByType: "install_deploys",
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
	}

	return nil
}

func (s *Signal) execApplyPlan(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, installWorkflowStepID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating deploy plan")
	l.Info("creating execution plan")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		return fmt.Errorf("unable to get build: %w", err)
	}

	// get previous job
	operation := app.RunnerJobOperationTypeCreateApplyPlan
	if installDeploy.Type == app.InstallDeployTypeTeardown {
		operation = app.RunnerJobOperationTypeCreateTeardownPlan
	}
	runnerJobType := build.ComponentConfigConnection.Type.DeployJobType()
	planJob, err := activities.AwaitGetLatestJob(ctx, &activities.GetLatestJobRequest{
		OwnerID:   installDeploy.ID,
		Operation: operation,
		Group:     runnerJobType.Group(),
		Type:      runnerJobType,
	})
	if err != nil {
		return errors.Wrap(err, "unable to plan runner job for current apply job")
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStreamID)
	}()

	op := app.RunnerJobOperationTypeApplyPlan
	jobTyp := build.ComponentConfigConnection.Type.DeployJobType()

	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID:    install.RunnerID,
		DeployID:    installDeploy.ID,
		Op:          op,
		Type:        jobTyp,
		LogStreamID: logStreamID,
		Metadata: map[string]string{
			"install_id":           install.ID,
			"deploy_id":            installDeploy.ID,
			"install_component_id": installDeploy.InstallComponentID,
			"component_id":         build.ComponentConfigConnection.ComponentID,
			"component_name":       build.ComponentConfigConnection.Component.Name,
			"operation":            string(op),
		},
	})
	if err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	// NOTE(jm): this is probably going to need to be refactored
	deployPlan, err := plan.AwaitCreateDeployPlan(ctx, &plan.CreateDeployPlanRequest{
		InstallDeployID: installDeploy.ID,
		InstallID:       install.ID,
		WorkflowID:      fmt.Sprintf("%s-create-apply-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create deploy plan")
		return errors.Wrap(err, "unable to create deploy plan")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         installWorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: "install_deploys",
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	// Add Plan contents from the result to the plan
	if runnerJob.Type == app.RunnerJobTypeJobNOOPDeploy {
		deployPlan.ApplyPlanContents = ""
		deployPlan.ApplyPlanDisplay = ""
	} else if len(planJob.Execution.Result.Contents) > 0 {
		l.Info("using the legacy contents from the runner job execution result")
		deployPlan.ApplyPlanContents = planJob.Execution.Result.Contents
		deployPlan.ApplyPlanDisplay = string(planJob.Execution.Result.ContentsDisplay)
	} else if len(planJob.Execution.Result.ContentsGzip) > 0 {
		l.Info("using the compressed contents from the runner job execution result")
		applyPlanContents, err := planJob.Execution.Result.GetContentsB64String()
		if err != nil {
			return errors.Wrap(err, "unable to get contents string")
		}
		deployPlan.ApplyPlanContents = applyPlanContents
		applyPlanContentsDisplay, err := planJob.Execution.Result.GetContentsDisplayString()
		if err != nil {
			return errors.Wrap(err, "unable to get contents display string")
		}
		deployPlan.ApplyPlanDisplay = applyPlanContentsDisplay
	}

	planJSON, err := json.Marshal(deployPlan)
	if err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create json from deploy plan")
		return errors.Wrap(err, "unable to create json from plan")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			DeployPlan: deployPlan,
		},
	}); err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	planJSON = nil
	deployPlan = nil

	s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "executing deploy plan")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   install.RunnerID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to execute runner job")
		l.Error("job did not succeed", zap.Error(err))
		return fmt.Errorf("unable to get install: %w", err)
	}

	return nil
}

func (s *Signal) updateDeployStatus(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string) {
	// TODO: GenerateStateRequest doesn't have DeployID field - needs InstallID
	// _, _ = installstate.AwaitGenerateState(ctx, &installstate.GenerateStateRequest{
	// 	InstallID: s.InstallID,
	// })

	// TODO: AwaitUpdateDeployStatusByDeployID removed - replace with AwaitUpdateDeployStatus
	// _ = activities.AwaitUpdateDeployStatusByDeployID(ctx, deployID, status, message)
}

func (s *Signal) updateDeployStatusWithoutStatusSync(ctx workflow.Context, deployID string, status app.InstallDeployStatus, message string) {
	// TODO: AwaitUpdateDeployStatusByDeployID removed - replace with AwaitUpdateDeployStatus
	// _ = activities.AwaitUpdateDeployStatusByDeployID(ctx, deployID, status, message)
}

func (s *Signal) updateInstallComponentStatus(ctx workflow.Context, installComponentID string, status app.InstallComponentStatus, message string) {
	// TODO: AwaitUpdateInstallComponentStatusByInstallComponentID removed - replace with AwaitUpdateInstallComponentStatus
	// _ = activities.AwaitUpdateInstallComponentStatusByInstallComponentID(ctx, installComponentID, status, message)
}
