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
// (from SignalWithMaxRetries) hasn't been exhausted, it creates a clone
// and writes a retry directive.
func (s *Signal) handleStepError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, stepErr error) error {
	sig := stepSignal(step)

	// Check auto-retry on inner signal
	ar, isAutoRetry := sig.(signal.SignalWithAutoRetry)
	if !isAutoRetry || !ar.AutoRetry() {
		// No auto-retry — mark error and return
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
		return stepErr
	}

	// Determine max retries from the signal, falling back to default
	maxRetries := signal.DefaultMaxRetries
	if mr, ok := sig.(signal.SignalWithMaxRetries); ok {
		maxRetries = mr.MaxRetries()
	}

	// Check if retries are exhausted before attempting clone
	nextRetryIndex := step.RetryIndex + 1
	if nextRetryIndex > maxRetries {
		l.Warn("max retries exhausted",
			zap.String("step_id", step.ID),
			zap.Int("max_retries", maxRetries),
			zap.Int("retry_index", step.RetryIndex))

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: flowutil.StepHumanDescription(stepErr),
				Metadata: map[string]any{
					"reason":            stepErr.Error(),
					"retries_exhausted": true,
					"max_retries":       maxRetries,
					"retry_index":       step.RetryIndex,
				},
			},
		}); err != nil {
			return errors.Wrap(err, "unable to mark step as error")
		}
		return stepErr
	}

	// Mark the step as failed with auto-retry metadata
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: flowutil.StepHumanDescription(stepErr),
			Metadata: map[string]any{
				"reason":       stepErr.Error(),
				"auto_retried": true,
				"retry_index":  step.RetryIndex,
				"max_retries":  maxRetries,
				DirectiveKey:   DirectiveRetry,
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step as error")
	}

	// Clone the step for retry
	if err := s.cloneWorkflowStep(ctx, step, flw); err != nil {
		l.Warn("auto-retry clone failed, returning original error",
			zap.String("step_id", step.ID),
			zap.Error(err))
		return stepErr
	}

	l.Debug("auto-retry triggered, cloned step",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID),
		zap.Int("retry_index", nextRetryIndex),
		zap.Int("max_retries", maxRetries))

	return nil
}
