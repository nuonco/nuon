package executeworkflowstepgroup

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// StepResult describes the outcome of executing a single step.
type StepResult struct {
	// Directive is the step's ResultDirective after execution. Empty or
	// "continue" means the caller should proceed to the next step.
	Directive string

	// Error is set when the step failed and was NOT handled by auto-retry or
	// user action. The caller should propagate this as a group error.
	Error error
}

// executeSingleStep dispatches a step, awaits its completion, and handles
// the full lifecycle: auto-retry (via directives), approval, user action on
// failure, and cancellation. Both executeSequential and executeParallel use
// this function.
//
// Returns a StepResult with the resolved directive. The caller decides how
// to map directives to group-level behavior (e.g. stop, retry-group, etc.).
func (s *Signal) executeSingleStep(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep) StepResult {
	l.Debug("dispatching step",
		zap.String("step_id", step.ID),
		zap.String("step_name", step.Name),
		zap.Int("group_idx", s.GroupIdx))

	// Dispatch the step signal.
	qsID, err := s.dispatchStep(ctx, step)
	if err != nil {
		l.Error("step dispatch error",
			zap.String("step_id", step.ID),
			zap.Error(err))
		return StepResult{Error: err}
	}

	// Track for cancellation.
	s.stepSignalIDs = append(s.stepSignalIDs, qsID)

	// Wait for the step to finish via the step-finished update handler.
	// This returns the step's final status and directive directly, avoiding
	// a separate DB re-fetch. It also survives handler termination+restart
	// via update-with-start.
	resp, err := activities.AwaitForwardStepFinished(ctx, activities.ForwardStepFinishedRequest{
		StepID: step.ID,
	})
	if err != nil {
		// Workflow-level cancellation.
		if ctx.Err() != nil {
			s.handleCancellation(ctx, l, step)
			return StepResult{Directive: DirectiveStop, Error: ctx.Err()}
		}

		// Step failed. The step signal may have auto-retried (returning nil
		// with a directive) — but if we got an error, auto-retry did NOT
		// trigger. Update workflow status and wait for user action.
		l.Debug("step failed, waiting for user action",
			zap.String("step_id", step.ID),
			zap.Error(err))

		_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.WorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "step failed, awaiting retry or skip",
				Metadata: map[string]any{
					"error_message":  err.Error(),
					"awaiting_retry": true,
					"step_id":        step.ID,
				},
			},
		})

		if awaitErr := s.awaitUserAction(ctx, l); awaitErr != nil {
			return StepResult{Error: awaitErr}
		}

		// Check if the user action triggered a retry-group.
		if s.retryGroupRequested {
			return StepResult{Directive: DirectiveRetryGroup}
		}

		// User action was handled (retry handler cloned, skip handler marked).
		// Return continue so the loop re-evaluates without double-cloning.
		return StepResult{Directive: DirectiveContinue}
	}

	// Step completed. Use the directive from the response directly.
	directive := resp.Directive
	l.Debug("step completed",
		zap.String("step_id", step.ID),
		zap.String("status", string(resp.Status)),
		zap.String("directive", directive))

	if directive == "" {
		if resp.Status == app.StatusError {
			directive = DirectiveStop
		} else {
			directive = DirectiveContinue
		}
	}

	return StepResult{Directive: directive}
}

// awaitUserAction blocks until the group receives a retry-step, skip-step, or
// cancel update. These update handlers modify the step state (clone, mark skipped)
// and then this method returns so the caller can re-evaluate.
func (s *Signal) awaitUserAction(ctx workflow.Context, l *zap.Logger) error {
	s.awaitingUserAction = true
	defer func() { s.awaitingUserAction = false }()

	l.Debug("group awaiting user action (retry-step, skip-step, or cancel)")

	if err := workflow.Await(ctx, func() bool {
		return s.userActionReceived || s.cancelRequested
	}); err != nil {
		return err
	}

	s.userActionReceived = false

	if s.cancelRequested {
		return errors.New("group cancelled while awaiting user action")
	}

	l.Debug("group received user action, resuming")
	return nil
}
