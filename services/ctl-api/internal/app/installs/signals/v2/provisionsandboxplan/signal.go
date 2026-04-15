package provisionsandboxplan

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "provision-sandbox-plan"

type Signal struct {
	InstallSandboxID string
	WorkflowStepID   string
	FlowStepID       string
	FlowID           string
	SandboxMode      bool

	cfg *internal.Config
}

var _ signal.Signal = &Signal{}
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		Operation: "sandbox-provision",
	}
}

func (s *Signal) WithParams(params *signal.Params) {
	s.cfg = params.Cfg
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
	s.FlowStepID = stepID
	s.FlowID = flowID
}

var _ signal.SignalWithStepContext = (*Signal)(nil)

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallSandboxID == "" {
		return fmt.Errorf("install sandbox id is required")
	}

	// Validate install sandbox exists
	_, err := activities.AwaitGetInstallForSandboxBySandboxID(ctx, s.InstallSandboxID)
	if err != nil {
		return fmt.Errorf("unable to get install for sandbox: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	install, err := activities.AwaitGetInstallForSandboxBySandboxID(ctx, s.InstallSandboxID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	installRun, err := activities.AwaitCreateSandboxRun(ctx, activities.CreateSandboxRunRequest{
		InstallID:  install.ID,
		RunType:    app.SandboxRunTypeProvision,
		WorkflowID: s.FlowID,
	})
	if err != nil {
		return fmt.Errorf("unable to create install: %w", err)
	}

	defer func() {
		if errors.Is(workflow.ErrCanceled, ctx.Err()) {
			updateCtx, updateCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer updateCtxCancel()
			s.updateRunStatusWithoutStatusSync(updateCtx, installRun.ID, app.SandboxRunStatusCancelled, "install sandbox run cancelled")
		}
	}()

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   installRun.ID,
			StepTargetType: "install_sandbox_runs",
		}); err != nil {
			return errors.Wrap(err, "unable to update install action workflow")
		}
	}

	defer func() {
		if pan := recover(); pan != nil {
			s.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "internal error")
			panic(pan)
		}
	}()

	s.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatus(app.InstallDeployStatusPlanning), "planning")

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		SandboxRunID: installRun.ID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)

	l := workflow.GetLogger(ctx)
	l.Info("executing provision plan")

	err = s.executeSandboxPlan(ctx, install, installRun, s.FlowStepID, s.SandboxMode, s.cfg.DNSRootDomain)
	if err != nil {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
		return err
	}

	s.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunPendingApproval, "pending approval")
	l.Info("provision plan was successful")
	return nil
}

func (s *Signal) executeSandboxPlan(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, stepID string, sandboxMode bool, dnsRootDomain string) error {
	l := workflow.GetLogger(ctx)

	op := app.RunnerJobOperationTypeCreateApplyPlan
	if installRun.RunType == app.SandboxRunTypeDeprovision {
		op = app.RunnerJobOperationTypeCreateTeardownPlan
	}

	runnerJob, err := activities.AwaitCreateSandboxJob(ctx, &activities.CreateSandboxJobRequest{
		InstallID: install.ID,
		RunnerID:  install.RunnerID,
		OwnerType: "install_sandbox_runs",
		OwnerID:   installRun.ID,
		Op:        op,
		Metadata: map[string]string{
			"install_id":       install.ID,
			"sandbox_run_id":   installRun.ID,
			"sandbox_run_type": string(installRun.RunType),
		},
	})
	if err != nil {
		s.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	planResponse, err := plan.AwaitCreateSandboxRunPlan(ctx, &plan.CreateSandboxRunPlanRequest{
		RunID:      installRun.ID,
		InstallID:  install.ID,
		RootDomain: dnsRootDomain,
		WorkflowID: fmt.Sprintf("%s-create-api-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install plan request")
		return errors.Wrap(err, "unable to create plan")
	}

	planJSON, err := json.Marshal(planResponse.Plan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			SandboxRunPlan: planResponse.Plan,
		},
		PermissionInfo: operationroles.NewPermissionInfo(planResponse.RoleSelection),
	}); err != nil {
		s.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "unable to save plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	// queue job
	l.Info("queued job and waiting on it to be picked up by runner event loop")
	status, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:    runnerJob.ID,
		RunnerID: install.RunnerID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		s.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "job failed")
		return fmt.Errorf("unable to execute job: %w", err)
	}
	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		s.updateRunStatusWithoutStatusSync(ctx, installRun.ID, app.SandboxRunStatusError, "job failed with status "+string(status))
		return fmt.Errorf("runner job failed with status %s", status)
	}

	job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job")
	}

	if _, err := activities.AwaitCreateStepApproval(ctx, &activities.CreateStepApprovalRequest{
		OwnerID:     installRun.ID,
		OwnerType:   "install_sandbox_runs",
		RunnerJobID: job.ID,
		StepID:      stepID,
		Type:        app.TerraformPlanApprovalType,
	}); err != nil {
		return errors.Wrap(err, "unable to create approval")
	}

	return nil
}

func (s *Signal) updateRunStatusWithoutStatusSync(ctx workflow.Context, runID string, status app.SandboxRunStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)

	if err := activities.AwaitUpdateRunStatus(ctx, activities.UpdateRunStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
		SkipStatusSync:    true,
	}); err != nil {
		l.Error("unable to update run status", zap.String("run-id", runID), zap.Error(err))
	}

	if err := statusactivities.AwaitUpdateRunStatusV2(ctx, statusactivities.UpdateRunStatusV2Request{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
	}); err != nil {
		l.Error("unable to update run status v2", zap.String("run-id", runID), zap.Error(err))
	}
}
