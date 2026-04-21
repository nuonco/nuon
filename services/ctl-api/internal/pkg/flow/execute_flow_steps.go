package flow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const (
	workflowStepBatchSize = 5
)

var FlowCancellationErr = fmt.Errorf("workflow cancelled")

func (c *WorkflowConductor[DomainSignal]) executeSteps(ctx workflow.Context, req eventloop.EventLoopRequest, flw *app.Workflow) error {
	return c.executeFlowSteps(ctx, req, flw, 0)
}

func (c *WorkflowConductor[DomainSignal]) executeFlowSteps(ctx workflow.Context, req eventloop.EventLoopRequest, flw *app.Workflow, startingStepNumber int) error {
	if flw.Status.Status == app.StatusCancelled {
		return FlowCancellationErr
	}

	steps := flw.Steps

	for i := startingStepNumber; i < len(steps); i++ {
		step := &steps[i]

		var err error
		if c.StepChildWorkflow {
			// Dispatch the full step lifecycle as a signal to the step queue.
			// The signal handles status updates, inner signal dispatch, approval,
			// policy checks, etc. Always re-fetch steps afterward since the signal
			// may have created clone steps or modified sibling steps.
			err = c.dispatchFlowStepSignal(ctx, step, flw)

			var refetchErr error
			steps, refetchErr = activities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, activities.GetFlowStepsRequest{
				FlowID: flw.ID,
			})
			if refetchErr != nil {
				return errors.Wrap(refetchErr, "unable to get steps after signal execution")
			}

			// Read directive from the executed step's metadata.
			// The step workflow writes directives to communicate how the parent should proceed.
			if err == nil {
				if directive := readStepDirective(steps, step.ID); directive != "" {
					switch directive {
					case DirectiveAwaitApproval:
						return NewApprovalPauseErr(step.ID)
					case DirectiveSkipGroup:
						i = findNextGroupStart(steps, step.GroupIdx, i)
						continue
					case DirectiveStop:
						if err := c.cancelFutureSteps(ctx, flw, i, "workflow step was denied"); err != nil {
							return errors.Wrap(err, "unable to cancel future steps")
						}
						return ErrNotApproved
					case DirectiveRetry:
						// Clone was already created by step workflow; re-fetch picked it up.
						// Don't advance i - the loop will pick up the clone at the next iteration.
						continue
					case DirectiveContinue:
						// Normal success, fall through to batch check
					}
				}
			}
		} else {
			var reFetchSteps bool
			reFetchSteps, err = c.executeFlowStep(ctx, req, step.Idx, step, flw)

			if reFetchSteps {
				// outer steps loop should continue to retry the step since the result here is ordered by idx asc
				steps, err = activities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, activities.GetFlowStepsRequest{
					FlowID: flw.ID,
				}) // this will re-query the steps from the database
				if err != nil {
					return errors.Wrap(err, "unable to get steps for retry")
				}
			}
		}

		if err == nil {
			nextIdx := i + 1
			if nextIdx < len(steps) && nextIdx-startingStepNumber == workflowStepBatchSize {
				// after a batch of steps, continue as new to avoid workflow history size limits
				return NewContinueAsNewErr(nextIdx)
			}
			continue
		}

		// handle cancellation
		if c.isCancellationErr(ctx, err) {
			return c.handleCancellation(ctx, err, step.ID, i, flw)
		}

		if errors.Is(err, ErrNotApproved) {
			if err := c.cancelFutureSteps(ctx, flw, i, "workflow step was not approved"); err != nil {
				return errors.Wrap(err, "unable to cancel future steps "+err.Error())
			}
			return err
		}

		// if the workflow was configured to abort, then go ahead and abort and do not attempt future steps
		if err := c.cancelFutureSteps(ctx, flw, i, "workflow step failed"); err != nil {
			return errors.Wrap(err, "unable to cancel future steps "+err.Error())
		}
		return err
	}

	return nil
}

// readStepDirective reads the directive from a step's status metadata.
func readStepDirective(steps []app.WorkflowStep, stepID string) string {
	for _, s := range steps {
		if s.ID == stepID {
			if s.Status.Metadata != nil {
				if d, ok := s.Status.Metadata[DirectiveKey]; ok {
					if ds, ok := d.(string); ok {
						return ds
					}
				}
			}
			return ""
		}
	}
	return ""
}

// findNextGroupStart returns the index of the first step that belongs to a different group
// than the given groupIdx, starting from afterIdx. If no such step exists, returns len(steps).
func findNextGroupStart(steps []app.WorkflowStep, groupIdx int, afterIdx int) int {
	for j := afterIdx + 1; j < len(steps); j++ {
		if steps[j].GroupIdx != groupIdx {
			return j - 1 // -1 because the loop will i++ to get to j
		}
	}
	return len(steps) // skip past all remaining steps
}
