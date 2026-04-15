package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// isCancellationErr delegates to IsCancellationErr.
func (c *WorkflowConductor[DomainSignal]) isCancellationErr(ctx workflow.Context, err error) bool {
	return IsCancellationErr(ctx, err)
}

// checkStepCancellation marks a step as cancelled if the workflow context is cancelled.
func (c *WorkflowConductor[DomainSignal]) checkStepCancellation(ctx workflow.Context, stepID string) error {
	return CheckStepCancellation(ctx, stepID)
}

// handleCancellation delegates to HandleCancellation.
func (c *WorkflowConductor[DomainSignal]) handleCancellation(ctx workflow.Context, stepErr error, stepID string, idx int, flw *app.Workflow) error {
	return HandleCancellation(ctx, stepErr, stepID, idx, flw)
}

// cancelFutureSteps delegates to CancelFutureSteps.
func (c *WorkflowConductor[DomainSignal]) cancelFutureSteps(ctx workflow.Context, flw *app.Workflow, idx int, reason string) error {
	return CancelFutureSteps(ctx, flw, idx, reason)
}
