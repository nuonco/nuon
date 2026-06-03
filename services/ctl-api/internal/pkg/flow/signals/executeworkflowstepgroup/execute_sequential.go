package executeworkflowstepgroup

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// executeSequential dispatches steps one at a time. After each step completes,
// the step's directive determines what happens next. Cloning for retry happens
// here — the single authoritative place for clone decisions.
func (s *Signal) executeSequential(ctx workflow.Context, l *zap.Logger) error {
	for {
		// Check if cancellation was requested before dispatching the next
		// step. This closes the race window where the cancel signal is
		// still propagating through the queue infrastructure but the
		// in-memory flag or the DB metadata already reflect it.
		if s.cancelRequested || s.isWorkflowCancelled(ctx) {
			s.cancelRequested = true
			return s.writeStepGroupDirective(ctx, directive.GroupStop)
		}

		steps, err := s.getGroupSteps(ctx)
		if err != nil {
			return err
		}

		step, found := s.nextExecutableStep(steps)
		if !found {
			return s.writeStepGroupDirective(ctx, directive.GroupContinue)
		}

		result := s.executeSingleStep(ctx, l, step)
		if result.Error != nil {
			return result.Error
		}

		r := result.Result
		switch r.Directive {
		case directive.StepContinue:
			continue

		case directive.StepRetry:
			// Clone the step for individual retry. The next iteration
			// picks up the pending clone.
			if err := CloneStepForRetry(ctx, step.ID, s.WorkflowID); err != nil {
				l.Warn("unable to clone step for retry", zap.String("step_id", step.ID), zap.Error(err))
				return err
			}
			continue

		case directive.StepStop:
			siblingStatus := r.SiblingStatus
			if siblingStatus == "" {
				siblingStatus = app.StatusDiscarded
			}
			s.cancelRemainingSteps(ctx, l, steps, step.ID, siblingStatus)
			return s.writeStepGroupDirective(ctx, directive.GroupStop)

		case directive.StepRetryGroup:
			s.cancelRemainingSteps(ctx, l, steps, step.ID, app.StatusDiscarded)
			return s.writeStepGroupDirective(ctx, directive.GroupRetryGroup)

		case directive.StepSkipGroup:
			siblingStatus := r.SiblingStatus
			if siblingStatus == "" {
				siblingStatus = app.StatusDiscarded
			}
			s.cancelRemainingSteps(ctx, l, steps, step.ID, siblingStatus)
			return s.writeStepGroupDirective(ctx, directive.GroupSkipGroup)

		case directive.StepAwaitApproval:
			return s.writeStepGroupDirective(ctx, directive.GroupAwaitApproval)

		default:
			continue
		}
	}
}
