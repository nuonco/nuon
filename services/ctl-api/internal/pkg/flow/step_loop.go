package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/flowutil"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// ExecuteStepsViaChildWorkflow runs the step iteration loop where each step is
// dispatched as an execute-workflow-step signal. This is the StepChildWorkflow
// mode used by the execute-flow signal.
func ExecuteStepsViaChildWorkflow(ctx workflow.Context, cfg StepConfig, flw *app.Workflow, startFromIdx int) error {
	if flw.Status.Status == app.StatusCancelled {
		return FlowCancellationErr
	}

	steps := flw.Steps

	for i := startFromIdx; i < len(steps); i++ {
		step := &steps[i]

		err := DispatchStepSignal(ctx, cfg, step, flw)

		var refetchErr error
		steps, refetchErr = activities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, activities.GetFlowStepsRequest{
			FlowID: flw.ID,
		})
		if refetchErr != nil {
			return errors.Wrap(refetchErr, "unable to get steps after signal execution")
		}

		// Read directive from the executed step's metadata.
		if err == nil {
			if directive := readStepDirective(steps, step.ID); directive != "" {
				switch directive {
				case flowutil.DirectiveAwaitApproval:
					return NewApprovalPauseErr(step.ID)
				case flowutil.DirectiveSkipGroup:
					i = findNextGroupStart(steps, step.GroupIdx, i)
					continue
				case flowutil.DirectiveStop:
					if err := CancelFutureSteps(ctx, flw, i, "workflow step was denied"); err != nil {
						return errors.Wrap(err, "unable to cancel future steps")
					}
					return ErrNotApproved
				case flowutil.DirectiveRetry:
					continue
				case flowutil.DirectiveContinue:
					// Normal success, fall through to batch check
				}
			}
		}

		if err == nil {
			nextIdx := i + 1
			if nextIdx < len(steps) && nextIdx-startFromIdx == workflowStepBatchSize {
				return NewContinueAsNewErr(nextIdx)
			}
			continue
		}

		// Handle cancellation
		if IsCancellationErr(ctx, err) {
			return HandleCancellation(ctx, err, step.ID, i, flw)
		}

		if errors.Is(err, ErrNotApproved) {
			if err := CancelFutureSteps(ctx, flw, i, "workflow step was not approved"); err != nil {
				return errors.Wrap(err, "unable to cancel future steps "+err.Error())
			}
			return err
		}

		// Abort on error — cancel future steps
		if err := CancelFutureSteps(ctx, flw, i, "workflow step failed"); err != nil {
			return errors.Wrap(err, "unable to cancel future steps "+err.Error())
		}
		return err
	}

	return nil
}
