package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/flowutil"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// handleStepError marks the step as errored and checks for auto-retry.
// If the inner signal implements SignalWithAutoRetry and the retry budget
// hasn't been exhausted, it creates a clone and writes a retry directive.
func (s *Signal) handleStepError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, stepErr error) error {
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status: app.StatusError,
			Metadata: map[string]any{
				"reason": stepErr.Error(),
			},
			StatusHumanDescription: flowutil.StepHumanDescription(stepErr),
		},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step as error")
	}

	// Check auto-retry on inner signal
	sig := stepSignal(step)
	if ar, ok := sig.(signal.SignalWithAutoRetry); ok && ar.AutoRetry() {
		if err := s.cloneWorkflowStep(ctx, step, flw); err != nil {
			l.Warn("auto-retry clone failed, returning original error",
				zap.String("step_id", step.ID),
				zap.Error(err))
			return stepErr
		}

		l.Debug("auto-retry triggered, cloned step",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))

		// Mark the original step as discarded since the clone will take over
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusDiscarded,
				StatusHumanDescription: "auto-retrying step",
				Metadata: map[string]any{
					"reason":     stepErr.Error(),
					DirectiveKey: DirectiveRetry,
				},
			},
		}); err != nil {
			l.Warn("failed to mark step as discarded for auto-retry", zap.Error(err))
		}

		return writeDirective(ctx, step.ID, DirectiveRetry, map[string]any{
			"step_idx": step.Idx,
			"status":   "auto-retrying",
		})
	}

	return stepErr
}
