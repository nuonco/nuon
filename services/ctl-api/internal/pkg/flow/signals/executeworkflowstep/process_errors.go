package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// handleStepCancelled stops the group when the step's inner signal reported a
// cancelled completion. Cancellation is never a failure: it must not enter the
// auto-retry path, and it must never let the group carry on to the next step.
// When cancellation came through Cancel() the directive and statuses are
// already written; an out-of-band cancellation (the inner queue signal was
// cancelled directly) writes them here.
func (s *Signal) handleStepCancelled(ctx workflow.Context, l *zap.Logger) error {
	if s.canceled {
		return nil
	}

	if err := setResultDirective(ctx, s.StepID, DirectiveStop); err != nil {
		return errors.Wrap(err, "unable to set stop directive for cancelled step")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCancelled,
			StatusHumanDescription: "step cancelled",
		},
	}); err != nil {
		l.Warn("failed to mark step as cancelled",
			zap.String("step_id", s.StepID),
			zap.Error(err))
	}

	return nil
}

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

	// Consult the composite-error hint recorded for this step's target. The
	// runner-result chokepoint parses a failed execution into a typed composite
	// error and mirrors its hints onto the target row before this step wakes.
	// A skip_auto_retry hint (e.g. a missing IAM permission that won't resolve
	// by retrying) forces the await-retry branch so we park for manual retry
	// instead of burning auto-retries. Only deploy/sandbox targets carry a
	// composite_error column, so we skip the lookup for every other target.
	skipAutoRetry := false
	if targetHasCompositeError(step.StepTargetType) {
		if hintsResp, herr := activities.AwaitGetStepErrorHints(ctx, activities.GetStepErrorHintsRequest{
			StepID: step.ID,
		}); herr != nil {
			l.Warn("unable to get step error hints",
				zap.String("step_id", step.ID),
				zap.Error(herr))
		} else if hintsResp != nil {
			skipAutoRetry = hintsResp.Hints.SkipAutoRetry()
		}
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

	// For retry-group signals the retry counter is GroupRetryIdx (reset per
	// group clone); for plain retry it is the step-level RetryIndex.
	retryGroup := false
	retryIndex := step.RetryIndex
	if rg, ok := sig.(signal.SignalWithRetryGroup); ok && rg.RetryGroup() {
		retryGroup = true
		retryIndex = step.GroupRetryIdx
	}

	nextRetryIndex := retryIndex + 1
	d := resolveFailureDirective(step.Skippable, retryGroup, skipAutoRetry, retryIndex, maxRetries, maxAutoRetries)

	// Skippable failure — mark the step as failed but let the workflow
	// continue. This allows post-trigger action steps to fail without
	// blocking the entire workflow.
	if d == DirectiveContinue {
		l.Info("step is skippable, continuing workflow after exhausted retries",
			zap.String("step_id", step.ID))
		meta := map[string]any{
			"max_retries":        maxRetries,
			"retry_index":        retryIndex,
			"skipped_on_failure": true,
		}
		if nextRetryIndex > maxRetries {
			meta["retries_exhausted"] = true
		} else {
			meta["auto_retries_exhausted"] = nextRetryIndex > maxAutoRetries
			meta["skip_auto_retry"] = skipAutoRetry
			meta["max_auto_retries"] = maxAutoRetries
		}
		_ = s.markStepFailed(ctx, step, stepErr, meta)
		if err := setResultDirective(ctx, step.ID, DirectiveContinue); err != nil {
			return errors.Wrap(err, "unable to set result directive")
		}
		return nil
	}

	// Global ceiling — no more retries of any kind.
	if d == DirectiveStop {
		l.Warn("max retries exhausted",
			zap.String("step_id", step.ID),
			zap.Int("max_retries", maxRetries),
			zap.Int("retry_index", retryIndex))

		if err := setResultDirective(ctx, step.ID, DirectiveStop); err != nil {
			return errors.Wrap(err, "unable to set result directive")
		}
		return s.markStepFailed(ctx, step, stepErr, map[string]any{
			"retries_exhausted": true,
			"max_retries":       maxRetries,
			"retry_index":       retryIndex,
		})
	}

	// Park for manual retry when auto-retries are exhausted OR the composite
	// error hinted that auto-retry won't help. The user can still manually
	// retry up to maxRetries.
	if d == DirectiveAwaitRetry {
		l.Warn("parking step for manual retry",
			zap.String("step_id", step.ID),
			zap.Bool("skip_auto_retry", skipAutoRetry),
			zap.Int("max_auto_retries", maxAutoRetries),
			zap.Int("max_retries", maxRetries),
			zap.Int("retry_index", retryIndex))

		// Mark step as errored and write the await-retry directive.
		// Legacy Execute() blocks here until the user retries or cancels.
		_ = s.markStepFailed(ctx, step, stepErr, map[string]any{
			"auto_retries_exhausted": nextRetryIndex > maxAutoRetries,
			"skip_auto_retry":        skipAutoRetry,
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

		if s.ResidentFlow {
			return nil
		}

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
		zap.String("directive", string(d)),
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
				DirectiveKey:   d,
			},
		},
	})

	// Write the directive. The group reads it and handles cloning.
	if err := setResultDirective(ctx, step.ID, d); err != nil {
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

// targetHasCompositeError reports whether a step's target type carries a
// composite_error column that the runner-result chokepoint mirrors hints onto.
// Only these targets are worth a GetStepErrorHints lookup.
//
// NOTE: Expand this as more target types gain composite_error columns.
func targetHasCompositeError(targetType string) bool {
	switch app.WorkflowStepTargetType(targetType) {
	case app.WorkflowStepTargetTypeInstallDeploy,
		app.WorkflowStepTargetTypeInstallDeploys,
		app.WorkflowStepTargetTypeInstallSandboxRun,
		app.WorkflowStepTargetTypeInstallSandboxRuns:
		return true
	default:
		return false
	}
}
