package executeworkflowstepgroup

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// StepResult describes the outcome of executing a single step.
type StepResult struct {
	// Result carries the directive and status metadata from the step.
	Result directive.StepResult

	// Error is set when the step failed unexpectedly (not handled by the
	// directive system). The caller should propagate this as a group error.
	Error error
}

// executeSingleStep dispatches a step, awaits its queue signal completion, and
// reads the step's directive from the database. Execute() stays alive until the
// directive is terminal (blocking for approval or retry), so AwaitQueueSignal
// naturally blocks for the full step lifecycle.
func (s *Signal) executeSingleStep(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep) StepResult {
	l.Debug("dispatching step",
		zap.String("step_id", step.ID),
		zap.String("step_name", step.Name),
		zap.Int("group_idx", s.GroupIdx))

	// Dispatch the step signal with a callback for completion notification.
	s.stepDispatchSeq++
	cb := callback.NewAttempt(ctx, step.ID, s.stepDispatchSeq)
	qsID, err := s.dispatchStep(ctx, step, cb)
	if err != nil {
		l.Error("step dispatch error",
			zap.String("step_id", step.ID),
			zap.Error(err))
		return StepResult{Error: err}
	}

	// Track for cancellation.
	s.stepSignalIDs = append(s.stepSignalIDs, qsID)

	// Await step completion. Execute() stays alive until the directive is
	// terminal, so this blocks for the full lifecycle including approval and
	// retry waiting. Bound by the step's derived timeout, falling back to the
	// human-wait cap when unset.
	stepTimeout := step.Timeout
	if stepTimeout <= 0 {
		stepTimeout = callback.FallbackAwaitTimeout
	}

	var qsErr error
	var updatedStep *app.WorkflowStep
	var d directive.Step
	for {
		_, qsErr = callback.AwaitWithTimeout(ctx, cb, stepTimeout)
		if ctx.Err() != nil {
			s.handleCancellation(ctx, l, step)
			return StepResult{
				Result: directive.NewStepResult(directive.StepStop),
				Error:  ctx.Err(),
			}
		}

		// Read the step's final state from DB.
		var err error
		updatedStep, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
		if err != nil {
			return StepResult{Error: err}
		}

		d = directive.Step(updatedStep.ResultDirective)

		// await-retry is non-terminal: the wait may have just timed out; keep
		// waiting so the step stays retryable instead of stopping.
		if d == directive.StepAwaitRetry {
			continue
		}
		break
	}

	l.Debug("step completed",
		zap.String("step_id", step.ID),
		zap.String("directive", string(d)))

	if qsErr != nil && d == "" {
		// Step failed without a directive — unexpected error.
		return StepResult{Error: qsErr}
	}

	if d == "" {
		d = directive.StepContinue
	}

	// Build the result with the step's status metadata for reason/status info.
	result := directive.NewStepResult(d)
	if updatedStep.Status.StatusHumanDescription != "" {
		result.Reason = updatedStep.Status.StatusHumanDescription
	}
	// Read optional status overrides from step metadata.
	if meta := updatedStep.Status.Metadata; meta != nil {
		if v, ok := meta["sibling_status"].(string); ok && v != "" {
			result.SiblingStatus = app.Status(v)
		}
		if v, ok := meta["future_step_status"].(string); ok && v != "" {
			result.FutureStatus = app.Status(v)
		}
	}

	return StepResult{Result: result}
}
