package flow

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"golang.org/x/net/context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CheckStepCancellation marks a step as cancelled if the workflow context is cancelled.
func CheckStepCancellation(ctx workflow.Context, stepID string) error {
	if errors.Is(ctx.Err(), workflow.ErrCanceled) {
		cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
		defer cancelCtxCancel()

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
			ID: stepID,
			Status: app.CompositeStatus{
				Status: app.StatusCancelled,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

// IsCancellationErr returns true if the error indicates workflow cancellation.
func IsCancellationErr(ctx workflow.Context, err error) bool {
	return errors.Is(ctx.Err(), workflow.ErrCanceled) || errors.Is(err, context.Canceled)
}

// HandleCancellation marks the current step and all future steps as cancelled.
func HandleCancellation(ctx workflow.Context, stepErr error, stepID string, idx int, flw *app.Workflow) error {
	if !IsCancellationErr(ctx, stepErr) {
		return stepErr
	}

	cancelCtx, _ := workflow.NewDisconnectedContext(ctx)

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(cancelCtx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            stepID,
		Status:            app.StatusCancelled,
		StatusDescription: "Cancelled",
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	if idx < len(flw.Steps)-1 {
		for _, cancelStep := range flw.Steps[idx+1:] {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
				ID: cancelStep.ID,
				Status: app.NewCompositeTemporalStatus(ctx, app.StatusNotAttempted, map[string]any{
					"reason": fmt.Sprintf("%s and workflow was configured with abort-on-error.", FlowCancellationErr.Error()),
				}),
			}); err != nil {
				return errors.Wrap(err, "unable to cancel step after cancelled workflow")
			}
		}
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
		ID: stepID,
		Status: app.NewCompositeTemporalStatus(ctx, app.StatusCancelled, map[string]any{
			"reason": FlowCancellationErr.Error(),
		}),
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow step as cancelled")
	}

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, flw.ID); err != nil {
		return errors.Wrap(err, "unable to set finished at on the workflow")
	}

	return stepErr
}

// CancelFutureSteps marks all steps after idx as not-attempted.
func CancelFutureSteps(ctx workflow.Context, flw *app.Workflow, idx int, reason string) error {
	if idx >= len(flw.Steps) {
		return nil
	}

	for _, cancelStep := range flw.Steps[idx+1:] {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: cancelStep.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusNotAttempted, map[string]any{
				"reason": fmt.Sprintf("%s and workflow was configured with abort-on-error.", reason),
			}),
		}); err != nil {
			return errors.Wrap(err, "unable to update status")
		}
	}

	return nil
}
