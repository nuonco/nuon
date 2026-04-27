package provisionsandboxapplyplan

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/provisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "provision-sandbox-apply-plan"

type Signal struct {
	InstallSandboxID  string
	FlowID            string
	FlowStepID        string
	InstallWorkflowID string
	SandboxMode       bool

	cfg         *internal.Config
	runnerJobID string
}

var _ signal.Signal = &Signal{}
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)
var _ signal.SignalWithAutoRetry = (*Signal)(nil)
var _ signal.SignalWithMaxRetries = (*Signal)(nil)
var _ signal.SignalWithMaxAutoRetries = (*Signal)(nil)

func (s *Signal) AutoRetry() bool                       { return true }
func (s *Signal) MaxRetries() int                       { return 5 }
func (s *Signal) MaxAutoRetries(_ workflow.Context) int { return 3 }

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	if s.runnerJobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, s.runnerJobID)
	}
	return nil
}

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
	s.FlowStepID = stepID
	s.FlowID = flowID
	s.InstallWorkflowID = flowID
}

var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithCloneSteps = (*Signal)(nil)

func (s *Signal) CloneSteps(originalStepName string) []signal.CloneStepDef {
	return []signal.CloneStepDef{
		{
			Signal: &provisionsandboxplan.Signal{
				InstallSandboxID: s.InstallSandboxID,
				SandboxMode:      s.SandboxMode,
			},
			Name:          originalStepName + " (plan)",
			ExecutionType: "approval",
		},
		{
			Signal: &Signal{
				InstallSandboxID: s.InstallSandboxID,
				SandboxMode:      s.SandboxMode,
			},
			Name:          originalStepName,
			ExecutionType: "system",
		},
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallSandboxID == "" {
		return fmt.Errorf("install sandbox id is required")
	}

	_, err := activities.AwaitGetInstallForSandboxBySandboxID(ctx, s.InstallSandboxID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	install, err := activities.AwaitGetInstallForSandboxBySandboxID(ctx, s.InstallSandboxID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	sandboxRun, err := activities.AwaitGetInstallSandboxRunForApplyStep(ctx, activities.GetInstallSandboxRunForApplyStep{
		InstallWorkflowID: s.FlowID,
		InstallID:         install.ID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get install deploy")
	}

	s.updateRunStatus(ctx, sandboxRun.ID, app.SandboxRunStatusProvisioning, "provisioning sandbox")

	ctx = cctx.SetLogStreamWorkflowContext(ctx, &sandboxRun.LogStream)
	l := workflow.GetLogger(ctx)

	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, sandboxRun.LogStream.ID)
	}()

	l.Info("executing plan")
	if err := s.executeApplyPlan(ctx, install, sandboxRun, s.FlowStepID, s.SandboxMode, s.cfg.DNSRootDomain); err != nil {
		s.updateRunStatus(ctx, sandboxRun.ID, app.SandboxRunStatusError, "job did not succeed")
		return errors.Wrap(err, "unable to execute deploy")
	}

	s.updateRunStatus(ctx, sandboxRun.ID, app.SandboxRunStatusActive, "successfully provisioned")
	if _, err := state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   s.InstallWorkflowID,
		TriggeredByType: "install_sandbox_runs",
	}); err != nil {
		l.Warn("unable to generate state", zap.Error(err))
	}
	return nil
}

func (s *Signal) executeApplyPlan(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, stepID string, sandboxMode bool, dnsRootDomain string) error {
	l := workflow.GetLogger(ctx)
	l.Info("executing apply plan")

	s.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatus(app.StatusApplying), "applying plan")

	operation := app.RunnerJobOperationTypeCreateApplyPlan
	if installRun.RunType == app.SandboxRunTypeDeprovision {
		operation = app.RunnerJobOperationTypeCreateTeardownPlan
	}
	planJob, err := activities.AwaitGetLatestJob(ctx, &activities.GetLatestJobRequest{
		OwnerID:   installRun.ID,
		Operation: operation,
		Group:     app.RunnerJobGroupSandbox,
		Type:      app.RunnerJobTypeSandboxTerraform,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get plan runner job for current apply job")
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStreamID)
	}()

	// create the job
	runnerJob, err := activities.AwaitCreateSandboxJob(ctx, &activities.CreateSandboxJobRequest{
		InstallID: install.ID,
		RunnerID:  install.RunnerID,
		OwnerType: "install_sandbox_runs",
		OwnerID:   installRun.ID,
		Op:        app.RunnerJobOperationTypeApplyPlan,
		Metadata: map[string]string{
			"install_id":       install.ID,
			"sandbox_run_id":   installRun.ID,
			"sandbox_run_type": string(installRun.RunType),
		},
		LogStreamID: logStreamID,
	})
	if err != nil {
		s.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}
	s.runnerJobID = runnerJob.ID

	l.Info("creating sandbox run plan")
	planResponse, err := plan.AwaitCreateSandboxRunPlan(ctx, &plan.CreateSandboxRunPlanRequest{
		RunID:      installRun.ID,
		InstallID:  install.ID,
		RootDomain: dnsRootDomain,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-api-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		s.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install plan request")
		return errors.Wrap(err, "unable to create plan")
	}

	if stepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         stepID,
			StepTargetID:   installRun.ID,
			StepTargetType: "install_sandbox_runs",
		}); err != nil {
			return errors.Wrap(err, "unable to update install workflow")
		}
	}

	// Add Plan contents from the result to the plan
	if len(planJob.Execution.Result.Contents) > 0 {
		l.Info("using the legacy contents from the runner job execution result")
		planResponse.Plan.ApplyPlanContents = planJob.Execution.Result.Contents
		planResponse.Plan.ApplyPlanDisplay = planJob.Execution.Result.ContentsDisplay
	} else if len(planJob.Execution.Result.ContentsGzip) > 0 {
		l.Info(
			"using the compressed contents from the runner job execution result",
			zap.Int("contents.bytes.compressed", len(planJob.Execution.Result.ContentsGzip)),
		)
		applyPlanContents, err := planJob.Execution.Result.GetContentsB64String()
		if err != nil {
			return errors.Wrap(err, "unable to get contents display string")
		}
		l.Info(
			"using the compressed contents from the runner job execution result",
			zap.Int("contents.bytes.compressed", len(planJob.Execution.Result.ContentsGzip)),
			zap.Int("contents.bytes.compressed.b64", len(applyPlanContents)),
		)
		planResponse.Plan.ApplyPlanContents = applyPlanContents
		applyPlanContentsDisplay, err := planJob.Execution.Result.GetContentsDisplayDecompressedBytes()
		if err != nil {
			return errors.Wrap(err, "unable to get contents display bytes")
		}
		planResponse.Plan.ApplyPlanDisplay = applyPlanContentsDisplay
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
	}); err != nil {
		s.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to save plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	if err := activities.AwaitRecordInstallRoleUsage(ctx, &activities.RecordInstallRoleUsageRequest{
		InstallID:     install.ID,
		RunnerJobID:   runnerJob.ID,
		RoleSelection: planResponse.RoleSelection,
	}); err != nil {
		s.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to record install role usage")
		return fmt.Errorf("unable to record install role usage: %w", err)
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
		s.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "job failed")
		return fmt.Errorf("unable to execute job: %w", err)
	}
	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		s.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "job failed with status"+string(status))
		return errors.New("job was not successful")
	}

	return nil
}

func (s *Signal) updateRunStatus(ctx workflow.Context, runID string, status app.SandboxRunStatus, statusDescription string) {
	l := workflow.GetLogger(ctx)

	if err := activities.AwaitUpdateRunStatus(ctx, activities.UpdateRunStatusRequest{
		RunID:             runID,
		Status:            status,
		StatusDescription: statusDescription,
		SkipStatusSync:    false,
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
