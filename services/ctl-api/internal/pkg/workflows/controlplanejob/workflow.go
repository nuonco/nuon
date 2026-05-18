package controlplanejob

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const TaskQueue = "control-plane-builds"

type Workflows struct{}

type ExecuteRequest struct {
	JobID string `json:"job_id" validate:"required"`
}

func NewWorkflows() *Workflows { return &Workflows{} }

func (w *Workflows) ExecuteControlPlaneJob(ctx workflow.Context, req *ExecuteRequest) error {
	ensureCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})
	execution, err := AwaitEnsureExecution(ensureCtx, &EnsureExecutionRequest{JobID: req.JobID})
	if err != nil {
		return fmt.Errorf("unable to ensure control-plane execution: %w", err)
	}

	runCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:           TaskQueue,
		StartToCloseTimeout: execution.JobExecutionTimeout,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	outcome := FinalizeOutcome{Success: true}
	if err := AwaitRunJob(runCtx, &RunJobRequest{JobID: req.JobID, ExecutionID: execution.ExecutionID}); err != nil {
		outcome.Success = false
		outcome.Error = err.Error()
	}

	finalizeCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
	if err := AwaitFinalize(finalizeCtx, &FinalizeRequest{JobID: req.JobID, ExecutionID: execution.ExecutionID, Outcome: outcome}); err != nil {
		return fmt.Errorf("unable to finalize control-plane execution: %w", err)
	}
	if !outcome.Success {
		return temporal.NewNonRetryableApplicationError(outcome.Error, "control-plane-build", fmt.Errorf("control-plane job failed: %s", outcome.Error))
	}
	return nil
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
