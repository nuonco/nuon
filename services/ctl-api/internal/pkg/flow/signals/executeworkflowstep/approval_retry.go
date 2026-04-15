package executeworkflowstep

import (
	"strconv"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// handleRetryResponse processes a "retry plan" response.
func (s *Signal) handleRetryResponse(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow) error {
	l.Debug("handling approval response type: retry plan",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	if or, ok := stepSignal(step).(signal.SignalWithOnRetry); ok {
		if err := or.OnRetry(ctx); err != nil {
			l.Warn("OnRetry hook failed", zap.Error(err))
		}
	}

	// Mark the current step as discarded
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusDiscarded,
			StatusHumanDescription: "retrying " + strconv.Itoa(step.Idx),
			Metadata: map[string]any{
				"step_idx":   step.Idx,
				"status":     "retrying",
				DirectiveKey: DirectiveRetry,
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to retry plan status")
	}

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            app.StatusDiscarded,
		StatusDescription: "Retrying step " + strconv.Itoa(step.Idx),
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	// Create the clone step for retry
	if err := s.cloneWorkflowStep(ctx, step, flw); err != nil {
		return errors.Wrap(err, "unable to clone step for retry")
	}

	return nil
}
