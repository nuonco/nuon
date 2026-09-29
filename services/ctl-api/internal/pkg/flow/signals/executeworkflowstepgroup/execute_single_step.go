package executeworkflowstepgroup

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type StepResult struct {
	Result directive.StepResult

	ManualRetry bool

	Error error
}

func (s *Signal) executeSingleStep(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep) StepResult {
	l.Debug("dispatching step",
		zap.String("step_id", step.ID),
		zap.String("step_name", step.Name),
		zap.Int("group_idx", s.GroupIdx))

	s.stepDispatchSeq++
	cb := callback.NewAttempt(ctx, step.ID, s.stepDispatchSeq)
	qsID, err := s.dispatchStep(ctx, step, cb)
	if err != nil {
		l.Error("step dispatch error",
			zap.String("step_id", step.ID),
			zap.Error(err))
		return StepResult{Error: err}
	}

	s.stepSignalIDs = append(s.stepSignalIDs, qsID)

	stepTimeout := step.Timeout
	if stepTimeout == 0 {
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

		var err error
		updatedStep, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
		if err != nil {
			return StepResult{Error: err}
		}

		d = directive.Step(updatedStep.ResultDirective)

		if d == directive.StepAwaitRetry && !s.ResidentFlow {
			continue
		}
		break
	}

	l.Debug("step completed",
		zap.String("step_id", step.ID),
		zap.String("directive", string(d)))

	if qsErr != nil && d == "" {
		return StepResult{Error: qsErr}
	}

	if d == "" {
		d = directive.StepContinue
	}

	result := directive.NewStepResult(d)
	manualRetry := false
	if updatedStep.Status.StatusHumanDescription != "" {
		result.Reason = updatedStep.Status.StatusHumanDescription
	}
	if meta := updatedStep.Status.Metadata; meta != nil {
		if retryType, ok := meta["retry_type"].(string); ok && retryType == "manual" {
			manualRetry = true
		}
		if v, ok := meta["sibling_status"].(string); ok && v != "" {
			result.SiblingStatus = app.Status(v)
		}
		if v, ok := meta["future_step_status"].(string); ok && v != "" {
			result.FutureStatus = app.Status(v)
		}
	}

	return StepResult{Result: result, ManualRetry: manualRetry}
}
