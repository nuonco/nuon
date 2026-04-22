package executeworkflowstepgroup

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// executeSequential dispatches steps one at a time, using executeSingleStep
// for the full step lifecycle (dispatch, await, auto-retry, user action).
func (s *Signal) executeSequential(ctx workflow.Context, l *zap.Logger) error {
	for {
		steps, err := s.getGroupSteps(ctx)
		if err != nil {
			return err
		}

		step, found := s.nextExecutableStep(steps)
		if !found {
			return s.writeStepGroupDirective(ctx, DirectiveContinue)
		}

		result := s.executeSingleStep(ctx, l, step)
		if result.Error != nil {
			return result.Error
		}

		switch result.Directive {
		case DirectiveContinue:
			continue

		case DirectiveRetry:
			// Auto-retry: the step signal wrote the directive but the group
			// owns cloning.
			if err := cloneStepForRetry(ctx, step.ID, s.WorkflowID); err != nil {
				l.Warn("unable to clone step for retry", zap.String("step_id", step.ID), zap.Error(err))
				return err
			}
			continue

		case DirectiveStop:
			s.cancelRemainingSteps(ctx, l, steps, step.ID, app.StatusDiscarded)
			return s.writeStepGroupDirective(ctx, DirectiveStop)

		case DirectiveRetryGroup:
			return s.writeStepGroupDirective(ctx, DirectiveRetryGroup)

		case DirectiveSkipGroup:
			s.cancelRemainingSteps(ctx, l, steps, step.ID, app.StatusDiscarded)
			return s.writeStepGroupDirective(ctx, DirectiveSkipGroup)

		case DirectiveAwaitApproval:
			return s.writeStepGroupDirective(ctx, DirectiveAwaitApproval)

		default:
			continue
		}
	}
}
