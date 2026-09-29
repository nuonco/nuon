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

type CancelStepRequest struct{}

type CancelStepResponse struct{}

func (s *Signal) cancelStepHandler(ctx workflow.Context, req CancelStepRequest) (*CancelStepResponse, error) {
	s.Cancel(ctx)
	return &CancelStepResponse{}, nil
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	// why: Flag first: the cancelled-execution path in Execute checks s.canceled
	// after any yield, so it must already be set when the directive write
	// below yields — otherwise the fallback writes a competing directive.
	s.canceled = true

	// why: we have to set the directive first, because if the workflow step returns, we don't want the group
	// to continue.
	// however, if we set the status before setting canceled to true, it does mean that the workflow could overwrite
	// it.
	setResultDirective(ctx, s.StepID, DirectiveStop)

	cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
	defer cancelCtxCancel()

	l, _ := log.WorkflowLogger(cancelCtx)

	if s.innerQueueSignalID != "" {
		_, err := client.AwaitCancelSignal(cancelCtx, s.innerQueueSignalID)
		if err != nil && l != nil {
			l.Warn("failed to cancel inner signal",
				zap.String("step_id", s.StepID),
				zap.String("inner_queue_signal_id", s.innerQueueSignalID),
				zap.Error(err))
		}
	}

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
