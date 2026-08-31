package controlplanejob

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const TaskQueue = "control-plane-builds"

type Workflows struct{}

type ExecuteRequest struct {
	JobID string `json:"job_id" validate:"required"`
}

func NewWorkflows() *Workflows { return &Workflows{} }

func (w *Workflows) ExecuteControlPlaneJob(ctx workflow.Context, req *ExecuteRequest) error {
	ensureCtx, _ := workflow.NewDisconnectedContext(ctx)
	ensureCtx = workflow.WithActivityOptions(ensureCtx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})
	execution, err := AwaitEnsureExecution(ensureCtx, EnsureExecutionRequest{JobID: req.JobID})
	if err != nil {
		return fmt.Errorf("unable to ensure control-plane execution: %w", err)
	}

	runCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           TaskQueue,
		StartToCloseTimeout: execution.JobExecutionTimeout,
		HeartbeatTimeout:    30 * time.Second,
		WaitForCancellation: true,
	})
	outcome := FinalizeOutcome{Status: app.RunnerJobExecutionStatusFinished}
	runErr := AwaitRunJob(
		runCtx,
		RunJobRequest{JobID: req.JobID, ExecutionID: execution.ExecutionID},
		&workflow.ActivityOptions{
			StartToCloseTimeout: execution.JobExecutionTimeout,
			RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
		},
	)
	if runErr != nil {
		outcome.Status = executionStatusForError(runErr)
		outcome.Error = runErr.Error()
	}
	if ctx.Err() != nil {
		outcome.Status = app.RunnerJobExecutionStatusCancelled
		outcome.Error = ctx.Err().Error()
		runErr = temporal.NewCanceledError()
	}

	finalizeCtx, _ := workflow.NewDisconnectedContext(ctx)
	finalizeCtx = workflow.WithActivityOptions(finalizeCtx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
	finalized, err := AwaitFinalize(finalizeCtx, FinalizeRequest{JobID: req.JobID, ExecutionID: execution.ExecutionID, Outcome: outcome})
	if err != nil {
		return fmt.Errorf("unable to finalize control-plane execution: %w", err)
	}
	if ctx.Err() != nil {
		return temporal.NewCanceledError()
	}
	if temporal.IsCanceledError(runErr) {
		return runErr
	}
	if finalized.Status == app.RunnerJobExecutionStatusCancelled {
		return temporal.NewCanceledError()
	}
	if finalized.Status != app.RunnerJobExecutionStatusFinished {
		message := outcome.Error
		if message == "" {
			message = string(finalized.Status)
		}
		return temporal.NewNonRetryableApplicationError(message, "control-plane-build", fmt.Errorf("control-plane job failed: %s", message))
	}
	return nil
}

func executionStatusForError(err error) app.RunnerJobExecutionStatus {
	if temporal.IsTimeoutError(err) {
		return app.RunnerJobExecutionStatusTimedOut
	}
	if temporal.IsCanceledError(err) {
		return app.RunnerJobExecutionStatusCancelled
	}
	return app.RunnerJobExecutionStatusFailed
}

func AwaitExecuteControlPlaneJob(ctx workflow.Context, req *ExecuteRequest, opts ...*workflow.ChildWorkflowOptions) error {
	cwo := workflow.ChildWorkflowOptions{}
	for _, opt := range opts {
		if opt != nil {
			cwo = *opt
		}
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	future := workflow.ExecuteChildWorkflow(ctx, (*Workflows).ExecuteControlPlaneJob, req)
	return future.Get(ctx, nil)
}
