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
// and writes a retry directive so the conductor continues to the clone.
func (s *Signal) handleStepError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, stepErr error) error {
	sig := stepSignal(step)

	// Check auto-retry on inner signal
	ar, isAutoRetry := sig.(signal.SignalWithAutoRetry)
	if !isAutoRetry || !ar.AutoRetry() {
		return s.markStepFailed(ctx, step, stepErr, nil)
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

		return s.markStepFailed(ctx, step, stepErr, map[string]any{
			"retries_exhausted": true,
			"max_retries":       maxRetries,
			"retry_index":       step.RetryIndex,
		})
	}

	// Clone the step first — only mark auto-retry metadata if clone succeeds
	if err := s.cloneWorkflowStep(ctx, step, flw); err != nil {
		l.Warn("auto-retry clone failed, returning original error",
			zap.String("step_id", step.ID),
			zap.Error(err))
		return s.markStepFailed(ctx, step, stepErr, map[string]any{
			"retries_exhausted": true,
			"clone_error":       err.Error(),
		})
	}

	l.Debug("auto-retry triggered, cloned step",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID),
		zap.Int("retry_index", nextRetryIndex),
		zap.Int("max_retries", maxRetries))

	// Clone succeeded — mark step as failed with auto-retry directive so the
	// conductor knows to continue to the cloned step.
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
		return errors.Wrap(err, "unable to mark step as auto-retried")
	}

	return nil
}

// markStepFailed writes a StatusError update for the step with the given error
// and optional extra metadata. It always returns stepErr.
func (s *Signal) markStepFailed(ctx workflow.Context, step *app.WorkflowStep, stepErr error, extraMeta map[string]any) error {
	meta := map[string]any{
		"reason": stepErr.Error(),
	}
	for k, v := range extraMeta {
		meta[k] = v
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: flowutil.StepHumanDescription(stepErr),
			Metadata:               meta,
		},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step as error")
	}
	return stepErr
}
