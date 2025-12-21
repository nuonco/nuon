package provisionsandboxapplyplan

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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const SignalType signal.SignalType = "provision-sandbox-apply-plan"

type Signal struct {
	InstallSandboxID  string
	FlowID            string
	FlowStepID        string
	InstallWorkflowID string
	SandboxMode       bool
	DNSRootDomain     string
}

var _ signal.Signal = &Signal{}

func (s *Signal) Type() signal.SignalType {
	return SignalType
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
	if err := s.executeApplyPlan(ctx, install, sandboxRun, s.FlowStepID, s.SandboxMode, s.DNSRootDomain); err != nil {
		s.updateRunStatus(ctx, sandboxRun.ID, app.SandboxRunStatusError, "job did not succeed")
		return errors.Wrap(err, "unable to execute deploy")
	}

	s.updateRunStatus(ctx, sandboxRun.ID, app.SandboxRunStatusActive, "successfully provisioned")
	_, err = state.AwaitGenerateState(ctx, &state.GenerateStateRequest{
		InstallID:       install.ID,
		TriggeredByID:   s.InstallWorkflowID,
		TriggeredByType: "install_sandbox_runs",
	})
	if err != nil {
		return errors.Wrap(err, "unable to generate state")
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

	l.Info("creating sandbox run plan")
	runPlan, err := plan.AwaitCreateSandboxRunPlan(ctx, &plan.CreateSandboxRunPlanRequest{
		RunID:      installRun.ID,
		InstallID:  install.ID,
		RootDomain: dnsRootDomain,
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
		runPlan.ApplyPlanContents = planJob.Execution.Result.Contents
		runPlan.ApplyPlanDisplay = planJob.Execution.Result.ContentsDisplay
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
		runPlan.ApplyPlanContents = applyPlanContents
		applyPlanContentsDisplay, err := planJob.Execution.Result.GetContentsDisplayDecompressedBytes()
		if err != nil {
			return errors.Wrap(err, "unable to get contents display bytes")
		}
		runPlan.ApplyPlanDisplay = applyPlanContentsDisplay
	}

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			SandboxRunPlan: runPlan,
		},
	}); err != nil {
		s.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to save plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	// queue job
	l.Info("queued job and waiting on it to be picked up by runner event loop")
	status, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:      runnerJob.ID,
		RunnerID:   install.RunnerID,
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
}
