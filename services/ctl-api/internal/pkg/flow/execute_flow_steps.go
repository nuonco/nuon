package flow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/flow/activities"
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

		reFetchSteps, err := c.executeFlowStep(ctx, req, step.Idx, step, flw)
		if reFetchSteps {
			// outer steps loop should continue to retry the step since the result here is ordered by idx asc
			steps, err = activities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, activities.GetFlowStepsRequest{
				FlowID: flw.ID,
			}) // this will re-query the steps from the database
			if err != nil {
				return errors.Wrap(err, "unable to get steps for retry")
			}
		}
		if err == nil {
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

		if flw.StepErrorBehavior == app.StepErrorBehaviorContinue {
			continue
		}

		// if the workflow was configured to abort, then go ahead and abort and do not attempt future steps
		if err := c.cancelFutureSteps(ctx, flw, i, "workflow step failed"); err != nil {
			return errors.Wrap(err, "unable to cancel future steps "+err.Error())
		}
		return err
	}

	return nil
}
