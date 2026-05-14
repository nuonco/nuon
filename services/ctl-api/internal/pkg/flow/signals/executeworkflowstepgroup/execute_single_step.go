package executeworkflowstepgroup

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// StepResult describes the outcome of executing a single step.
type StepResult struct {
	// Directive is the step's ResultDirective after execution.
	Directive directive.Step

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

	// Await step completion. Execute() stays alive until the directive is
	// terminal, so this blocks for the full lifecycle including approval
	// waiting and retry waiting.
	_, qsErr := client.AwaitQueueSignal(ctx, qsID)
	if ctx.Err() != nil {
		s.handleCancellation(ctx, l, step)
		return StepResult{Directive: directive.StepStop, Error: ctx.Err()}
	}

	// Read the step's final directive from DB.
	updatedStep, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return StepResult{Error: err}
	}

	d := directive.Step(updatedStep.ResultDirective)
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

	return StepResult{Directive: d}
}
