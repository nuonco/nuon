package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

// handleStepError marks the step as errored and checks for auto-retry.
//
// Step-error processing happens in three layers:
//
//  1. If the underlying error carries an stderr.ErrUser (either as a direct
//     wrap or via the stderr.StepErrorPayload attached as ApplicationError
//     details by AwaitSignal), and that ErrUser has a non-default Directive,
//     the directive determines the outcome:
//     - StepDirectiveStop → mark error, no auto-retry, surface user copy.
//     - StepDirectiveSkip → mark step as skipped (Success+continue), advance.
//
//  2. Otherwise, if the signal implements SignalWithAutoRetry and the retry
//     budget hasn't been exhausted, write a "retry" / "retry-group"
//     directive and return nil. The group reads the directive and clones.
//
//  3. Otherwise, mark the step failed and return the error.
func (s *Signal) handleStepError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, stepErr error) error {
	if u, ok := stderr.ExtractUserError(stepErr); ok {
		switch u.Directive {
		case stderr.StepDirectiveStop:
			return s.markStepStopped(ctx, l, step, u, stepErr)
		case stderr.StepDirectiveSkip:
			return s.markStepSkippedFromUserError(ctx, l, step, u)
		}
		// StepDirectiveDefault and any future directive falls through to
		// the existing auto-retry policy below.
	}

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
			zap.String("directive", directive),
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

	// Check auto-retry budget — user can still manually retry up to maxRetries.
	if nextRetryIndex > maxAutoRetries {
		l.Warn("auto retries exhausted",
			zap.String("step_id", step.ID),
			zap.Int("max_auto_retries", maxAutoRetries),
			zap.Int("max_retries", maxRetries),
			zap.Int("retry_index", retryIndex))

		return s.markStepFailed(ctx, step, stepErr, map[string]any{
			"auto_retries_exhausted": true,
			"max_auto_retries":       maxAutoRetries,
			"max_retries":            maxRetries,
			"retry_index":            retryIndex,
		})
	}

	l.Debug("auto-retry: writing directive",
		zap.String("step_id", step.ID),
		zap.String("directive", directive),
		zap.Int("retry_index", nextRetryIndex),
		zap.Int("max_retries", maxRetries))

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

// markStepStopped marks the step as errored with directive=Stop, skipping
// the auto-retry path entirely. The user-facing copy on the step status is
// taken from u.Description so the dashboard renders the curated message
// rather than the raw cause.
func (s *Signal) markStepStopped(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, u stderr.ErrUser, stepErr error) error {
	l.Info("step error has Stop directive — skipping auto-retry",
		zap.String("step_id", step.ID),
		zap.String("reason_code", u.Code))

	if err := setResultDirective(ctx, step.ID, DirectiveStop); err != nil {
		return errors.Wrap(err, "unable to set stop directive")
	}

	desc := u.Description
	if desc == "" {
		desc = stepHumanDescription(stepErr)
	}

	meta := mergeUserMetadata(map[string]any{
		"reason":   stepErr.Error(),
		"terminal": true,
	}, u)

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: desc,
			Metadata:               meta,
		},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step as stopped")
	}
	return stepErr
}

// markStepSkippedFromUserError marks the step as skipped with directive=Continue,
// matching the existing single-step skip semantics used by approval_skip.go:
// the step status is StatusSuccess with metadata.status="skipped". Returns
// nil so the group continues to the next step.
func (s *Signal) markStepSkippedFromUserError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, u stderr.ErrUser) error {
	l.Info("step error has Skip directive — marking step skipped",
		zap.String("step_id", step.ID),
		zap.String("reason_code", u.Code))

	extra := map[string]any{
		"step_idx": step.Idx,
		"status":   "skipped",
	}
	if u.Description != "" {
		extra["description"] = u.Description
	}
	extra = mergeUserMetadata(extra, u)

	return writeDirective(ctx, step.ID, DirectiveContinue, extra)
}

// mergeUserMetadata copies the StepErrorPayload-derived metadata
// (error_code / error_fields / step_directive) from u into base in place
// and returns it. base must be non-nil. Existing keys in base are
// preserved on collision (the caller's seed takes precedence).
func mergeUserMetadata(base map[string]any, u stderr.ErrUser) map[string]any {
	for k, v := range stderr.MetadataFromErrUser(u) {
		if _, exists := base[k]; exists {
			continue
		}
		base[k] = v
	}
	return base
}
