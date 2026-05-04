package executeworkflowstep

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// waitForRetryBackoff blocks until the step's RetryNotBeforeAt time is reached
// or the user requests an immediate retry via the "retry-now" update.
//
// While waiting, the step status is set to StatusWaitingToRetry with metadata
// describing the deadline so the dashboard can render a countdown. When the
// gate releases, the status is updated to in-progress and Execute proceeds
// with normal step execution.
//
// Returns nil immediately when:
//   - The step has no RetryNotBeforeAt set (typical first attempt or manual retry).
//   - The deadline has already passed (e.g. handler restart after the wait).
//   - The user signals retry-now.
//   - The context is cancelled (canceled / cancel-step).
func (s *Signal) waitForRetryBackoff(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep) error {
	if step.RetryNotBeforeAt == nil {
		return nil
	}

	now := workflow.Now(ctx)
	remaining := step.RetryNotBeforeAt.Sub(now)
	if remaining <= 0 {
		return nil
	}

	l.Debug("step waiting for auto-retry backoff",
		zap.String("step_id", step.ID),
		zap.Time("retry_not_before_at", *step.RetryNotBeforeAt),
		zap.Duration("remaining", remaining))

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusWaitingToRetry,
			StatusHumanDescription: "waiting to auto-retry after backoff",
			Metadata: map[string]any{
				"retry_not_before_at": step.RetryNotBeforeAt.Format(time.RFC3339),
				"backoff_seconds":     int64(remaining.Seconds()),
				"retry_idx":           step.RetryIndex,
			},
		},
	}); err != nil {
		// Non-fatal: status is informational. Continue with the wait.
		l.Warn("unable to mark step as waiting-to-retry",
			zap.String("step_id", step.ID),
			zap.Error(err))
	}

	// Wait until either the deadline passes, the user signals retry-now, or
	// the workflow is cancelled. AwaitWithTimeout returns true if the
	// condition fired before the timeout.
	awoke, err := workflow.AwaitWithTimeout(ctx, remaining, func() bool {
		return s.retryNowRequested || s.canceled
	})
	if err != nil {
		// Workflow cancellation propagates as an error here.
		return nil
	}

	if awoke && s.retryNowRequested {
		l.Debug("retry-now signal received, breaking out of backoff",
			zap.String("step_id", step.ID))
	}

	return nil
}
