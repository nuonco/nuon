package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// parkedWaitCeilingVersion gates the parked-step wait ceiling: histories
// written before the ceiling recorded an unbounded Await (no timer command),
// so replaying them with AwaitWithTimeout is nondeterministic.
const parkedWaitCeilingVersion = "parked-step-wait-ceiling-v1"

// targetlessStepCompositeErrorVersion gates the hints lookup for steps with no
// step target: histories written before it have no such activity command.
const targetlessStepCompositeErrorVersion = "targetless-step-composite-error-v1"

const terminalTargetlessStopVersion = "terminal-targetless-step-stop-v1"

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
		stepCE := s.targetlessStepCompositeError(ctx, l, step)
		if stepCE != nil &&
			stepCE.Hints.Terminal() &&
			workflow.GetVersion(ctx, terminalTargetlessStopVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
			if err := s.updateStepFailedStatus(ctx, step, stepErr, nil, stepCE); err != nil {
				return err
			}
			if err := setResultDirective(ctx, step.ID, DirectiveStop); err != nil {
				return errors.Wrap(err, "unable to set stop directive")
			}
			return nil
		}
		return s.markStepFailed(ctx, step, stepErr, nil, stepCE)
	}

	// Consult the composite-error hint recorded for this step's target. The
	// runner-result chokepoint parses a failed execution into a typed composite
	// error before this step wakes.
	// A skip_auto_retry hint (e.g. a missing IAM permission that won't resolve
	// by retrying) forces the await-retry branch so we park for manual retry
	// instead of burning auto-retries.
	skipAutoRetry := false
	terminal := false
	var stepCE *compositeerrors.CompositeErrorData
	if targetSupportsCompositeErrorHints(step.StepTargetType) {
		if hintsResp, herr := activities.AwaitGetStepErrorHints(ctx, activities.GetStepErrorHintsRequest{
			StepID: step.ID,
		}); herr != nil {
			l.Warn("unable to get step error hints",
				zap.String("step_id", step.ID),
				zap.Error(herr))
		} else if hintsResp != nil {
			skipAutoRetry = hintsResp.Hints.SkipAutoRetry()
			terminal = hintsResp.Hints.Terminal()
			stepCE = hintsResp.Error
		}
	}

	// A terminal failure cannot succeed on any retry, so parking for a manual
	// one would offer the user an action guaranteed to fail. Stop instead, and
	// let the reason on the step explain why.
	if terminal {
		l.Warn("step failed terminally, not retrying",
			zap.String("step_id", step.ID))

		metadata := map[string]any{"terminal": true}
		directive := DirectiveStop
		if step.SkipOnFailure {
			metadata["skipped_on_failure"] = true
			directive = DirectiveContinue
		}

		if err := setResultDirective(ctx, step.ID, directive); err != nil {
			return errors.Wrap(err, "unable to set result directive")
		}
		return s.markStepFailed(ctx, step, stepErr, metadata, stepCE)
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
	d := resolveFailureDirective(step.SkipOnFailure, retryGroup, skipAutoRetry, retryIndex, maxRetries, maxAutoRetries)

	// Skip-on-failure — mark the step as failed but let the workflow
	// continue. This allows post-trigger action steps to fail without
	// blocking the entire workflow.
	if d == DirectiveContinue {
		l.Info("step is skip-on-failure, continuing workflow after exhausted retries",
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
		_ = s.markStepFailed(ctx, step, stepErr, meta, stepCE)
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
		}, stepCE)
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
		}, stepCE)
		if err := setResultDirective(ctx, step.ID, DirectiveAwaitRetry); err != nil {
			return errors.Wrap(err, "unable to set await-retry directive")
		}

		// Park the workflow as failed-pending-retry: the dashboard, cancel
		// guards, and completion gates all key off this status. The
		// awaiting_retry flag marks the row as manually recoverable.
		_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: flw.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusFailedPendingRetry,
				StatusHumanDescription: "step failed, awaiting retry or skip",
				Metadata: map[string]any{
					"step_id":        step.ID,
					"awaiting_retry": true,
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
		// The ceiling stops abandoned parks from holding Temporal workflows
		// open forever.
		if workflow.GetVersion(ctx, parkedWaitCeilingVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
			return workflow.Await(ctx, func() bool { return s.retried || s.canceled || s.skipped })
		}
		parked, err := workflow.AwaitWithTimeout(ctx, callback.MaxWaitCeiling, func() bool {
			return s.retried || s.canceled || s.skipped
		})
		if err != nil {
			return err
		}
		if !parked {
			abandonedDesc := abandonedHumanDescription(stepErr)
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: abandonedDesc,
					CompositeError:         stepCE,
					Metadata: map[string]any{
						"abandoned":      true,
						"original_error": stepErr.Error(),
					},
				},
			}); err != nil {
				return errors.Wrap(err, "unable to update workflow step status")
			}
			if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
				StepID:            step.ID,
				Status:            app.StatusError,
				StatusDescription: abandonedDesc,
			}); err != nil {
				return errors.Wrap(err, "unable to update step target status for abandoned step")
			}
			if derr := setResultDirective(ctx, step.ID, DirectiveStop); derr != nil {
				return errors.Wrap(derr, "unable to set stop directive for abandoned step")
			}
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
			CompositeError:         stepCE,
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

// markStepFailed writes a StatusError update for the step with the given error,
// the parsed composite error when one was recorded for the step's target, and
// optional extra metadata. It always returns stepErr.
func (s *Signal) markStepFailed(ctx workflow.Context, step *app.WorkflowStep, stepErr error, extraMeta map[string]any, stepCE *compositeerrors.CompositeErrorData) error {
	if err := s.updateStepFailedStatus(ctx, step, stepErr, extraMeta, stepCE); err != nil {
		return err
	}
	return stepErr
}

func (s *Signal) updateStepFailedStatus(ctx workflow.Context, step *app.WorkflowStep, stepErr error, extraMeta map[string]any, stepCE *compositeerrors.CompositeErrorData) error {
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
			CompositeError:         stepCE,
			Metadata:               meta,
		},
	}); err != nil {
		return errors.Wrap(err, "unable to mark step as error")
	}
	return nil
}

// targetlessStepCompositeError reads back the composite error a targetless step
// (e.g. an app branch run step) recorded on its own status, so marking the step
// failed does not clear it.
func (s *Signal) targetlessStepCompositeError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep) *compositeerrors.CompositeErrorData {
	if step.StepTargetID != "" || step.StepTargetType != "" {
		return nil
	}
	if workflow.GetVersion(ctx, targetlessStepCompositeErrorVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return nil
	}

	hintsResp, err := activities.AwaitGetStepErrorHints(ctx, activities.GetStepErrorHintsRequest{
		StepID: step.ID,
	})
	if err != nil {
		l.Warn("unable to get step error hints",
			zap.String("step_id", step.ID),
			zap.Error(err))
		return nil
	}
	if hintsResp == nil {
		return nil
	}
	return hintsResp.Error
}

func targetSupportsCompositeErrorHints(targetType string) bool {
	switch app.WorkflowStepTargetType(targetType) {
	case app.WorkflowStepTargetTypeInstallDeploy,
		app.WorkflowStepTargetTypeInstallDeploys,
		app.WorkflowStepTargetTypeInstallSandboxRun,
		app.WorkflowStepTargetTypeInstallSandboxRuns,
		app.WorkflowStepTargetTypeInstallStackVersions:
		return true
	default:
		return false
	}
}
