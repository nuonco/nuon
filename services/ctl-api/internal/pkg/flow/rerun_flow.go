package flow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type RerunOperation string

const (
	RerunOperationSkipStep  RerunOperation = "skip-step"
	RerunOperationRetryStep RerunOperation = "retry-step"
)

type RerunInput struct {
	FlowID          string         `json:"flow_id" validate:"required"`
	StepID          string         `json:"step_id" validate:"required"`
	Operation       RerunOperation `json:"operation" validate:"required"`
	StalePlan       bool           `json:"stale_plan"`
	RePlanStepID    string         `json:"replan_step_id"`
	ContinueFromIdx int            `json:"continue_from_idx"`
}

// updateFlowStatusWithError is a helper function to update flow status with error information
func (c *WorkflowConductor[SignalType]) updateFlowStatusWithError(
	ctx workflow.Context,
	flowID string,
	description string,
	err error,
) error {
	return statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: description,
			Metadata: map[string]any{
				"error_message": err.Error(),
			},
		},
	})
}

// updateFlowStatus is a helper function to update flow status
func (c *WorkflowConductor[SignalType]) updateFlowStatus(
	ctx workflow.Context,
	flowID string,
	status app.Status,
	description string,
	metadata map[string]any,
) error {
	return statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flowID,
		Status: app.CompositeStatus{
			Status:                 status,
			StatusHumanDescription: description,
			Metadata:               metadata,
		},
	})
}

// Rerun is a workflow that reruns a flow from a specific step.
// It marks the existing step as discarded and creates a new step with the same parameters.
// It then executes the flow steps from the newly created step.
func (c *WorkflowConductor[SignalType]) Rerun(ctx workflow.Context, req eventloop.EventLoopRequest, inp RerunInput) error {
	// generate steps
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil
	}

	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, inp.FlowID)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow object")
	}

	if flw.Status.Status == app.StatusCancelled {
		return errors.New("workflow already cancelled")
	}

	defer func() {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()

			if err := c.updateFlowStatus(
				cancelCtx, flw.ID, app.StatusCancelled, "", nil); err != nil {
				l.Error("unable to update status on cancellation", zap.Error(err))
			}
		}
	}()

	defer func() {
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, inp.FlowID); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	var workflowStartStepNumber int
	if inp.ContinueFromIdx == 0 {
		flw, workflowStartStepNumber, err = c.prepareWorkflowForRerun(ctx, inp, flw)
		if err != nil {
			return errors.Wrap(err, "unable to prepare workflow for rerun")
		}
	} else {
		// We are continuing as a new (temporal) workflow, so we don't need to
		// prepare the workflow and duplicate steps again. Just fetch the steps and
		// start from the provided index.
		workflowStartStepNumber = inp.ContinueFromIdx

		flowSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, inp.FlowID)
		if err != nil {
			if err := c.updateFlowStatusWithError(
				ctx, inp.FlowID, "unable to fetch workflow step", err); err != nil {
				return errors.Wrap(err, "unable to update flow status with error")
			}
			return errors.Errorf("unable to fetch workflow steps for workflow %s: %v", inp.FlowID, err)
		}

		flw.Steps = make([]app.WorkflowStep, len(flowSteps))
		for i, step := range flowSteps {
			flw.Steps[i] = app.WorkflowStep(step)
		}
	}

	l.Debug("re-executing steps for workflow", zap.String("workflow_id", flw.ID), zap.Int("start_step_number", workflowStartStepNumber))
	if err := c.executeFlowSteps(ctx, req, flw, workflowStartStepNumber); err != nil {
		_, ok := err.(*ContinueAsNewErr)
		if ok {
			return err
		}

		if err := c.updateFlowStatus(
			ctx, inp.FlowID, app.StatusError, "error while executing steps",
			map[string]any{"error_message": err.Error()},
		); err != nil {
			return err
		}

		return errors.Wrap(err, "unable to execute workflow steps")
	}

	if err := c.updateFlowStatus(
		ctx, inp.FlowID, app.StatusSuccess, "successfully executed workflow", nil); err != nil {
		return err
	}

	return nil
}

func (c *WorkflowConductor[SignalType]) prepareWorkflowForRerun(ctx workflow.Context, inp RerunInput, flw *app.Workflow) (*app.Workflow, int, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, 0, nil
	}

	// reset state of the flow
	if err := c.updateFlowStatus(
		ctx, inp.FlowID, app.StatusRetrying, "", nil); err != nil {
		l.Error("unable to update status on retry", zap.Error(err))
	}

	if err := activities.AwaitPkgWorkflowsFlowResetFlowFinishedAtByID(ctx, inp.FlowID); err != nil {
		l.Error("unable to reset finished at", zap.Error(err))
	}

	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, inp.StepID)
	if err != nil {
		if err := c.updateFlowStatusWithError(
			ctx, inp.FlowID, "unable to fetch workflow step", err); err != nil {
			return nil, 0, err
		}
		return nil, 0, errors.Errorf("unable to fetch a step for workflow %s: %v", inp.FlowID, err)
	}

	// update the status of retryig step to discarded
	var stepStatusHumanDescription string
	var status app.Status
	var reason string

	switch inp.Operation {
	case RerunOperationRetryStep:
		stepStatusHumanDescription = "Step deployment failed."
		status = app.StatusDiscarded
		reason = "The step was discarded and retried by the user."
	case RerunOperationSkipStep:
		stepStatusHumanDescription = "Step skipped, continuing with next step."
		status = app.StatusUserSkipped
		reason = "The step was skipped by the user."
	default:
		err := fmt.Errorf("invalid rerun step operation %s", inp.Operation)
		if err := c.updateFlowStatusWithError(ctx, inp.FlowID, err.Error(), err); err != nil {
			return nil, 0, err
		}
		return nil, 0, err
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 status,
			StatusHumanDescription: stepStatusHumanDescription,
			Metadata: map[string]any{
				"reason": reason,
			},
		},
	}); err != nil {
		if err := c.updateFlowStatusWithError(
			ctx, inp.FlowID, stepStatusHumanDescription, errors.New(reason)); err != nil {
			return nil, 0, err
		}

		return nil, 0, errors.Wrapf(err, "unable to update flow step %s status to discarded", step.ID)
	}

	if err := c.updateFlowStatus(
		ctx, inp.FlowID, app.StatusInProgress, "generating steps for flow", nil); err != nil {
		l.Error("unable to update status on retry", zap.Error(err))
	}

	l.Debug("generating steps for flow")
	if inp.Operation == RerunOperationRetryStep {
		updatedGroupRetryIdx := step.GroupRetryIdx

		// if current plan has a stale plan, create new plan step
		// in case of stale plan, retry entire group for now it includes plan + apply step
		if inp.StalePlan {
			l.Debug("retry step is a apply plan wiht stale plan")
			l.Debug("updating group number for new plan set and new step group")
			updatedGroupRetryIdx += 1

			planStep, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, inp.RePlanStepID)
			if err != nil {
				if err := c.updateFlowStatusWithError(
					ctx, inp.FlowID, "unable to fetch replan step for rerun apply step", err); err != nil {
					return nil, 0, err
				}
				return nil, 0, errors.Errorf(
					"unable to fetch replan steps %s for apply step %s for workflow %s: %v",
					inp.RePlanStepID, inp.RePlanStepID, inp.FlowID, err,
				)
			}

			// this should never be the case ideally, only here for initial validation
			// can be removed once planstep id is coming in correctly
			if step.GroupIdx != planStep.GroupIdx {
				return nil, 0, errors.Errorf(
					"invalid plan step for input retry step, groupIdx for retry step %s is %d, groupIdx for stale plan step %s is %d",
					step.ID, step.GroupIdx, planStep.ID, planStep.GroupIdx,
				)
			}

			l.Debug("updating status for stale plan step")
			// update the status of plan step to discarded
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: planStep.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusDiscarded,
					StatusHumanDescription: "Plan step discarded, continuing with retry.",
					Metadata: map[string]any{
						"reason": "The plan step was discarded and retried by the user.",
					},
				},
			}); err != nil {
				if err := c.updateFlowStatusWithError(
					ctx, inp.FlowID, "unable to update plan step status to discarded", err); err != nil {
					return nil, 0, err
				}
				return nil, 0, errors.Wrapf(err, "unable to update plan step %s status to discarded", planStep.ID)
			}

			// fix grop idx and retry count for plan step since this will be in new group
			planStep.GroupRetryIdx = updatedGroupRetryIdx
			planStep.Name = removeRetryFromStepName(planStep.Name)
			step.Name = removeRetryFromStepName(step.Name)

			l.Debug("creating new plan step")
			if err = c.cloneWorkflowStep(ctx, planStep, flw); err != nil {
				if err := c.updateFlowStatusWithError(
					ctx, inp.FlowID, "unable to create plan retry step for apply step", err); err != nil {
					return nil, 0, err
				}
				return nil, 0, errors.Wrapf(err, "unable to create retry plan step for step %s workflow %s", inp.StepID, inp.FlowID)
			}
		}

		// create new retry step
		step.GroupRetryIdx = updatedGroupRetryIdx
		l.Debug("creating new retry plan")
		if err := c.cloneWorkflowStep(ctx, step, flw); err != nil {
			if err := c.updateFlowStatusWithError(
				ctx, inp.FlowID, "unable to create retry step", err); err != nil {
				return nil, 0, err
			}
			return nil, 0, errors.Wrapf(err, "unable to create retry step for workflow %s", inp.FlowID)
		}
	}

	flowSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, inp.FlowID)
	if err != nil {
		if err := c.updateFlowStatusWithError(
			ctx, inp.FlowID, "unable to fetch workflow steps", err); err != nil {
			return nil, 0, err
		}
		return nil, 0, errors.Errorf("unable to fetch steps for workflow %s: %v", inp.FlowID, err)
	}

	// get the index of newly created step
	var workflowStartStepNumber int
	for i, step := range flowSteps {
		if step.ID == inp.StepID {
			// if the step was retried it'll start from the new retry step
			// if the step was not retried, it'll start from next step
			// if apply step was retried, it'll start from retrying the plan step then apply step
			workflowStartStepNumber = i + 1
			break
		}
	}

	for _, s := range flowSteps[workflowStartStepNumber:] {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.ID,
			Status: app.CompositeStatus{
				Status:   app.StatusPending,
				Metadata: map[string]any{"reason": ""},
			},
		}); err != nil {
			return nil, 0, errors.Wrap(err, "unable to update status")
		}
	}

	if err := c.updateFlowStatus(
		ctx, inp.FlowID, app.StatusInProgress, "successfully generated all steps", nil); err != nil {
		return nil, 0, err
	}

	// re-fetch steps to get the steps with updated statuses. Current state is stale and results in
	// the steps being skipped over as they are not marked as pending.
	flw.Steps, err = activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, inp.FlowID)
	if err != nil {
		if err := c.updateFlowStatusWithError(
			ctx, inp.FlowID, "unable to fetch workflow steps", err); err != nil {
			return nil, 0, err
		}
		return nil, 0, errors.Errorf("unable to fetch steps for workflow %s: %v", inp.FlowID, err)
	}

	return flw, workflowStartStepNumber, nil
}
