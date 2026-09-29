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

const parkedWaitCeilingVersion = "parked-step-wait-ceiling-v1"

const targetlessStepCompositeErrorVersion = "targetless-step-composite-error-v1"

const terminalTargetlessStopVersion = "terminal-targetless-step-stop-v1"

// why: handleStepCancelled stops the group when the step's inner signal reported a
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

func (s *Signal) handleStepError(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, stepErr error) error {
	sig := stepSignal(step)

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

	// why: A terminal failure cannot succeed on any retry, so parking for a manual
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

	maxRetries := signal.DefaultMaxRetries
	if mr, ok := sig.(signal.SignalWithMaxRetries); ok {
		maxRetries = mr.MaxRetries()
	}

	maxAutoRetries := maxRetries
	if mar, ok := sig.(signal.SignalWithMaxAutoRetries); ok {
		maxAutoRetries = mar.MaxAutoRetries(ctx)
	}

	retryGroup := false
	retryIndex := step.RetryIndex
	if rg, ok := sig.(signal.SignalWithRetryGroup); ok && rg.RetryGroup() {
		retryGroup = true
		retryIndex = step.GroupRetryIdx
	}

	nextRetryIndex := retryIndex + 1
	d := resolveFailureDirective(step.SkipOnFailure, retryGroup, skipAutoRetry, retryIndex, maxRetries, maxAutoRetries)

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

	if d == DirectiveAwaitRetry {
		l.Warn("parking step for manual retry",
			zap.String("step_id", step.ID),
			zap.Bool("skip_auto_retry", skipAutoRetry),
			zap.Int("max_auto_retries", maxAutoRetries),
			zap.Int("max_retries", maxRetries),
			zap.Int("retry_index", retryIndex))

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

	if or, ok := sig.(signal.SignalWithOnRetry); ok {
		if err := or.OnRetry(ctx); err != nil {
			l.Warn("OnRetry hook failed", zap.String("step_id", step.ID), zap.Error(err))
		}
	}

	// why: Record auto-retry metadata on the error status. We intentionally do NOT
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

	if err := setResultDirective(ctx, step.ID, d); err != nil {
		return errors.Wrap(err, "unable to set result directive")
	}

	return nil
}

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
