package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// handleStepError marks the step as errored and checks for auto-retry.
// If the inner signal implements SignalWithAutoRetry and the retry budget
// hasn't been exhausted, it writes a directive ("retry" or "retry-group")
// and returns nil. The group reads the directive and handles cloning.
func (s *Signal) handleStepError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, stepErr error) error {
	sig := stepSignal(step)

	// Check auto-retry on inner signal.
	ar, isAutoRetry := sig.(signal.SignalWithAutoRetry)
	if !isAutoRetry || !ar.AutoRetry() {
		return s.markStepFailed(ctx, step, stepErr, nil)
	}

	// Determine max retries from the signal, falling back to default.
	maxRetries := signal.DefaultMaxRetries
	if mr, ok := sig.(signal.SignalWithMaxRetries); ok {
		maxRetries = mr.MaxRetries()
	}

	// Determine max auto-retries. Defaults to maxRetries when not implemented.
	maxAutoRetries := maxRetries
	if mar, ok := sig.(signal.SignalWithMaxAutoRetries); ok {
		maxAutoRetries = mar.MaxAutoRetries(ctx)
	}

	// Determine the directive based on signal capabilities. For retry-group
	// signals the retry counter is GroupRetryIdx (reset per group clone);
	// for plain retry it is the step-level RetryIndex.
	directive := DirectiveRetry
	retryIndex := step.RetryIndex
	if rg, ok := sig.(signal.SignalWithRetryGroup); ok && rg.RetryGroup() {
		directive = DirectiveRetryGroup
		retryIndex = step.GroupRetryIdx
	}

	nextRetryIndex := retryIndex + 1

	// Check the global ceiling first — no more retries of any kind.
	if nextRetryIndex > maxRetries {
		l.Warn("max retries exhausted",
			zap.String("step_id", step.ID),
			zap.String("directive", string(directive)),
			zap.Int("max_retries", maxRetries),
			zap.Int("retry_index", retryIndex))

		// If the step is skippable, mark it as failed but continue the workflow
		// instead of stopping. This allows post-trigger action steps to fail
		// without blocking the entire workflow.
		if step.Skippable {
			l.Info("step is skippable, continuing workflow after exhausted retries",
				zap.String("step_id", step.ID))
			_ = s.markStepFailed(ctx, step, stepErr, map[string]any{
				"retries_exhausted":  true,
				"max_retries":        maxRetries,
				"retry_index":        retryIndex,
				"skipped_on_failure": true,
			})
			if err := setResultDirective(ctx, step.ID, DirectiveContinue); err != nil {
				return errors.Wrap(err, "unable to set result directive")
			}
			return nil
		}

		if err := setResultDirective(ctx, step.ID, DirectiveStop); err != nil {
			return errors.Wrap(err, "unable to set result directive")
		}
		return s.markStepFailed(ctx, step, stepErr, map[string]any{
			"retries_exhausted": true,
			"max_retries":       maxRetries,
			"retry_index":       retryIndex,
		})
	}

	// Check auto-retry budget — user can still manually retry up to maxRetries.
	if nextRetryIndex > maxAutoRetries {
		l.Warn("auto retries exhausted",
			zap.String("step_id", step.ID),
			zap.Int("max_auto_retries", maxAutoRetries),
			zap.Int("max_retries", maxRetries),
			zap.Int("retry_index", retryIndex))

		// If the step is skippable and all retries (including manual) are exhausted
		// at the auto-retry level, allow the workflow to continue.
		if step.Skippable && maxAutoRetries >= maxRetries {
			l.Info("step is skippable, continuing workflow after exhausted retries",
				zap.String("step_id", step.ID))
			_ = s.markStepFailed(ctx, step, stepErr, map[string]any{
				"auto_retries_exhausted": true,
				"max_auto_retries":       maxAutoRetries,
				"max_retries":            maxRetries,
				"retry_index":            retryIndex,
				"skipped_on_failure":     true,
			})
			if err := setResultDirective(ctx, step.ID, DirectiveContinue); err != nil {
				return errors.Wrap(err, "unable to set result directive")
			}
			return nil
		}

		// Mark step as errored and write the await-retry directive.
		// Execute() blocks here until the user retries or cancels.
		_ = s.markStepFailed(ctx, step, stepErr, map[string]any{
			"auto_retries_exhausted": true,
			"max_auto_retries":       maxAutoRetries,
			"max_retries":            maxRetries,
			"retry_index":            retryIndex,
		})
		if err := setResultDirective(ctx, step.ID, DirectiveAwaitRetry); err != nil {
			return errors.Wrap(err, "unable to set await-retry directive")
		}

		// Update workflow status so the UI shows "failed pending retry".
		_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: flw.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusFailedPendingRetry,
				StatusHumanDescription: "step failed, awaiting retry or skip",
				Metadata: map[string]any{
					"step_id": step.ID,
				},
			},
		})

		// Block until user retries or cancels. The group's AwaitQueueSignal
		// stays blocked naturally. When the retry update arrives (flow → group → step),
		// s.retried is set and we unblock. The createStepRetryHandler writes the
		// terminal directive (retry or retry-group) before setting s.retried.
		if err := workflow.Await(ctx, func() bool { return s.retried || s.canceled || s.skipped }); err != nil {
			return err
		}
		return nil
	}

	l.Debug("auto-retry: writing directive",
		zap.String("step_id", step.ID),
		zap.String("directive", string(directive)),
		zap.Int("retry_index", nextRetryIndex),
		zap.Int("max_retries", maxRetries))

	// Call OnRetry so the signal can update its target object's status
	// (e.g. mark a deploy or sandbox run as "retried").
	if or, ok := sig.(signal.SignalWithOnRetry); ok {
		if err := or.OnRetry(ctx); err != nil {
			l.Warn("OnRetry hook failed", zap.String("step_id", step.ID), zap.Error(err))
		}
	}

	// Record auto-retry metadata on the error status. We intentionally do NOT
	// set retried=true here — the dashboard uses that flag to hide the error,
	// and we want the error to remain visible. The auto_retried metadata field
	// is sufficient to indicate this step was automatically retried.
	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: stepHumanDescription(stepErr),
			Metadata: map[string]any{
				"reason":       stepErr.Error(),
				"auto_retried": true,
				"retry_type":   "auto",
				"retry_idx":    retryIndex,
				"max_retries":  maxRetries,
				DirectiveKey:   directive,
			},
		},
	})

	// Write the directive. The group reads it and handles cloning.
	if err := setResultDirective(ctx, step.ID, directive); err != nil {
		return errors.Wrap(err, "unable to set result directive")
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
			StatusHumanDescription: stepHumanDescription(stepErr),
			Metadata:               meta,
		},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step as error")
	}
	return stepErr
}
