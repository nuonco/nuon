package flow

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"golang.org/x/net/context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (c *WorkflowConductor[DomainSignal]) isCancellationErr(ctx workflow.Context, err error) bool {
	if errors.Is(ctx.Err(), workflow.ErrCanceled) || errors.Is(err, context.Canceled) {
		return true
	}

	return false
}

func (c *WorkflowConductor[DomainSignal]) checkStepCancellation(ctx workflow.Context, stepID string) error {
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

func (c *WorkflowConductor[DomainSignal]) handleCancellation(ctx workflow.Context, stepErr error, stepID string, idx int, flw *app.Workflow) error {
	if !c.isCancellationErr(ctx, stepErr) {
		return stepErr
	}

	cancelCtx, _ := workflow.NewDisconnectedContext(ctx)

	// update the step target
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(cancelCtx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            stepID,
		Status:            app.StatusCancelled,
		StatusDescription: "Cancelled",
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	// cancel all steps
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

	// update the workflow status
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
		ID: stepID,
		Status: app.NewCompositeTemporalStatus(ctx, app.StatusCancelled, map[string]any{
			"reason": FlowCancellationErr.Error(),
		}),
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow step as cancelled")
	}

	// close the workflow
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, flw.ID); err != nil {
		return errors.Wrap(err, "unable to set finished at on the workflow")
	}

	return stepErr
}

func (c *WorkflowConductor[DomainSignal]) cancelFutureSteps(ctx workflow.Context, flw *app.Workflow, idx int, reason string) error {
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
