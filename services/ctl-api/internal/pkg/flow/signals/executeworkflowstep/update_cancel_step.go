package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CancelStepRequest is the input for the "cancel-step" update handler.
type CancelStepRequest struct{}

// CancelStepResponse is the response from the "cancel-step" update handler.
type CancelStepResponse struct{}

// cancelStepHandler delegates to the signal's Cancel() method which propagates
// cancellation to the inner (target) signal and updates step status.
func (s *Signal) cancelStepHandler(ctx workflow.Context, req CancelStepRequest) (*CancelStepResponse, error) {
	s.Cancel(ctx)
	return &CancelStepResponse{}, nil
}

// Cancel propagates cancellation to the inner signal if one is currently executing.
// The inner signal's Cancel() method is responsible for updating the underlying
// resource (e.g. stack run) status to cancelled.
func (s *Signal) Cancel(ctx workflow.Context) error {
	// NOTE(jm): we have to set the directive first, because if the workflow step returns, we don't want the group
	// to continue.
	// however, if we set the status before setting canceled to true, it does mean that the workflow could overwrite
	// it.
	setResultDirective(ctx, s.StepID, DirectiveStop)

	s.canceled = true

	cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
	defer cancelCtxCancel()

	l, _ := log.WorkflowLogger(cancelCtx)

	if s.innerQueueSignalID != "" {
		// Cancel the inner signal — this propagates to the actual step signal
		// (e.g. sandbox plan, deploy) which should update its resource status.
		_, err := client.AwaitCancelSignal(cancelCtx, s.innerQueueSignalID)
		if err != nil && l != nil {
			l.Warn("failed to cancel inner signal",
				zap.String("step_id", s.StepID),
				zap.String("inner_queue_signal_id", s.innerQueueSignalID),
				zap.Error(err))
		}
	}

	// Mark the step as cancelled
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status: app.StatusCancelled,
		},
	}); err != nil && l != nil {
		l.Warn("failed to mark step as cancelled",
			zap.String("step_id", s.StepID),
			zap.Error(err))
	}

	// Update the step target status to cancelled
	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(cancelCtx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            s.StepID,
		Status:            app.StatusCancelled,
		StatusDescription: "Cancelled",
	}); err != nil && l != nil {
		l.Warn("failed to update step target status on cancel",
			zap.String("step_id", s.StepID),
			zap.Error(err))
	}

	return nil
}
