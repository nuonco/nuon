package componentteardownapplyplan

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/stategen"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentteardownsyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "component-teardown-apply-plan"

type Signal struct {
	signal.LifecycleBase

	InstallComponentID    string
	InstallID             string
	ComponentID           string
	FlowID                string
	FlowStepID            string
	InstallWorkflowStepID string
	SandboxMode           bool

	runnerJobID string
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.InstallWorkflowStepID = stepID
	s.FlowStepID = stepID
	s.FlowID = flowID
}

var (
	_ signal.SignalWithStepContext      = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithCloneSteps       = (*Signal)(nil)
	_ signal.SignalWithCancel           = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
	_ signal.SignalWithMaxAutoRetries   = (*Signal)(nil)
	_ signal.SignalWithOnRetry          = (*Signal)(nil)
)

func (s *Signal) OnRetry(ctx workflow.Context) error {
	deploy, err := activities.AwaitGetInstallDeployForApplyStep(ctx, activities.GetInstallDeployForApplyStep{
		InstallWorkflowID: s.FlowID,
		ComponentID:       s.ComponentID,
	})
	if err != nil {
		return nil
	}
	s.updateDeployStatus(ctx, deploy.ID, app.InstallDeployStatusRetried, "deploy retried")
	return nil
}

func (s *Signal) AutoRetry() bool { return true }
func (s *Signal) MaxRetries() int { return 15 }

func (s *Signal) MaxAutoRetries(ctx workflow.Context) int {
	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return 0
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return 0
	}

	for _, ccc := range appCfg.ComponentConfigConnections {
		if ccc.ComponentID == s.ComponentID {
			return ccc.GetMaxAutoRetries()
		}
	}

	return 0
}

func (s *Signal) Clone(_ workflow.Context, originalStepName string) ([]signal.CloneStepDef, error) {
	return []signal.CloneStepDef{
		{
			Signal: &componentteardownsyncandplan.Signal{
				InstallComponentID: s.InstallComponentID,
				InstallID:          s.InstallID,
				ComponentID:        s.ComponentID,
				SandboxMode:        s.SandboxMode,
			},
			Name:          originalStepName + " (plan)",
			ExecutionType: "approval",
		},
		{
			Signal: &Signal{
				InstallComponentID: s.InstallComponentID,
				InstallID:          s.InstallID,
				ComponentID:        s.ComponentID,
				SandboxMode:        s.SandboxMode,
			},
			Name:          originalStepName,
			ExecutionType: "system",
		},
	}, nil
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	if s.runnerJobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, s.runnerJobID)
	}
	return nil
}

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	installID := &s.InstallID
	if s.InstallID == "" {
		installID = nil
	}
	componentID := &s.ComponentID
	if s.ComponentID == "" {
		componentID = nil
	}
	return signal.SignalLifecycleContext{
		InstallID:    installID,
		ComponentID:  componentID,
		Operation:    "component-teardown",
		Stage:        "apply",
		WorkflowID:   s.LifecycleWorkflowID,
		WorkflowType: s.LifecycleWorkflowType,
	}
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

	_ = activities.AwaitSetDeployAppliedAt(ctx, activities.SetDeployAppliedAtRequest{
		DeployID: installDeploy.ID,
	})

	s.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusInactive, "successfully torn down")
	s.updateInstallComponentStatus(ctx, installDeploy.InstallComponentID, app.InstallComponentStatusInactive, "successfully torn down")

	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	if err := stategen.HintOrGenerate(ctx, stategen.Request{
		StateGenV2:      statemanager.UseStateGenV2(orgEnabled, install.Metadata),
		InstallID:       install.ID,
		Targets:         statemanager.TargetsForHint(statemanager.HintComponentTeardown, s.InstallComponentID),
		ForceAll:        true,
		TriggeredByID:   installDeploy.ID,
		TriggeredByType: "install_deploys",
	}); err != nil {
		return err
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
	s.runnerJobID = runnerJob.ID

	// NOTE(jm): this is probably going to need to be refactored
	deployPlan, err := plan.AwaitCreateDeployPlan(ctx, &plan.CreateDeployPlanRequest{
		InstallDeployID: installDeploy.ID,
		InstallID:       install.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-apply-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
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
		deployPlan.Plan.ApplyPlanContents = ""
		deployPlan.Plan.ApplyPlanDisplay = ""
	} else if len(planJob.Execution.Result.Contents) > 0 {
		l.Info("using the legacy contents from the runner job execution result")
		deployPlan.Plan.ApplyPlanContents = planJob.Execution.Result.Contents
		deployPlan.Plan.ApplyPlanDisplay = string(planJob.Execution.Result.ContentsDisplay)
	} else if len(planJob.Execution.Result.ContentsGzip) > 0 {
		l.Info("using the compressed contents from the runner job execution result")
		applyPlanContents, err := planJob.Execution.Result.GetContentsB64String()
		if err != nil {
			return errors.Wrap(err, "unable to get contents string")
		}
		deployPlan.Plan.ApplyPlanContents = applyPlanContents
		applyPlanContentsDisplay, err := planJob.Execution.Result.GetContentsDisplayString()
		if err != nil {
			return errors.Wrap(err, "unable to get contents display string")
		}
		deployPlan.Plan.ApplyPlanDisplay = applyPlanContentsDisplay
	}

	planJSON, err := json.Marshal(deployPlan.Plan)
	if err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create json from deploy plan")
		return errors.Wrap(err, "unable to create json from plan")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			DeployPlan: deployPlan.Plan,
		},
	}); err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     install.ID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: deployPlan.RoleSelection,
	}); err != nil {
		s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to record install role usage")
		return fmt.Errorf("unable to record install role usage: %w", err)
	}

	planJSON = nil

	s.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "executing deploy plan")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID: install.RunnerID,
		JobID:    runnerJob.ID,
	}, &workflow.ChildWorkflowOptions{
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
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateDeployStatus(ctx, activities.UpdateDeployStatusRequest{
		DeployID:          deployID,
		Status:            status,
		StatusDescription: message,
		SkipStatusSync:    false,
	}); err != nil {
		l.Error("unable to update deploy status",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:          deployID,
		Status:            app.Status(status),
		StatusDescription: message,
		SkipStatusSync:    false,
	}); err != nil {
		l.Error("unable to update deploy status v2",
			zap.String("deploy-id", deployID),
			zap.Error(err))
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
		l.Error("unable to update deploy status",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:          deployID,
		Status:            app.Status(status),
		StatusDescription: message,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update deploy status v2",
			zap.String("deploy-id", deployID),
			zap.Error(err))
	}
}

func (s *Signal) updateInstallComponentStatus(ctx workflow.Context, installComponentID string, status app.InstallComponentStatus, message string) {
	l := workflow.GetLogger(ctx)
	if err := activities.AwaitUpdateInstallComponentStatus(ctx, activities.UpdateInstallComponentStatusRequest{
		InstallComponentID: installComponentID,
		Status:             status,
		StatusDescription:  message,
	}); err != nil {
		l.Error("unable to update install component status",
			zap.String("InstallComponentID", installComponentID),
			zap.Error(err))
	}
}
