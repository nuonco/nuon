package flow

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/flow/activities"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

type RerunInput struct {
	FlowID    string `json:"flow_id" validate:"required"`
	StepID    string `json:"step_id" validate:"required"`
	RetryStep bool   `json:"retry_step"`
}

// Rerun is a workflow that reruns a flow from a specific step.
// It marks the existing step as discarded and creates a new step with the same parameters.
// It then executes the flow steps from the newly created step.
func (c *FlowConductor[SignalType]) Rerun(ctx workflow.Context, req eventloop.EventLoopRequest, inp RerunInput) error {
	// generate steps
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil
	}

	flw, err := activities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, inp.FlowID)
	if err != nil {
		return errors.Wrap(err, "unable to get flow object")
	}

	if flw.Status.Status == app.StatusCancelled {
		return errors.New("flow already cancelled")
	}

	// reset state of the flow
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: inp.FlowID,
		Status: app.CompositeStatus{
			Status: app.StatusRetrying,
		},
	}); err != nil {
		l.Error("unable to update status on retry", zap.Error(err))
	}

	if err := activities.AwaitPkgWorkflowsFlowResetFlowFinishedAtByID(ctx, inp.FlowID); err != nil {
		l.Error("unable to reset finished at", zap.Error(err))
	}

	defer func() {
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()

			if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(cancelCtx, statusactivities.UpdateStatusRequest{
				ID: flw.ID,
				Status: app.CompositeStatus{
					Status: app.StatusCancelled,
				},
			}); err != nil {
				l.Error("unable to update status on cancellation", zap.Error(err))
			}
		}
	}()

	defer func() {
		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, inp.FlowID); err != nil {
			l.Error("unable to update finished at", zap.Error(err))
		}
	}()

	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, inp.StepID)
	if err != nil {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: inp.FlowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "unable to fetch flow step",
				Metadata: map[string]any{
					"error_message": err.Error(),
				},
			},
		}); err != nil {
			return err
		}
		return errors.Errorf("unable to fetch flow steps for flow %s: %v", inp.FlowID, err)
	}

	// update the status of retryig step to discarded
	var stepStatusHumanDescription string
	var status app.Status
	var reason string
	if inp.RetryStep {
		stepStatusHumanDescription = "Step deployment failed."
		status = app.StatusDiscarded
		reason = "The step was discarded and retried by the user."
	} else {
		stepStatusHumanDescription = "Step skipped, continuing with next step."
		status = app.StatusUserSkipped
		reason = "The step was skipped by the user."
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
		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: inp.FlowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: stepStatusHumanDescription,
				Metadata: map[string]any{
					"reason": reason,
				},
			},
		}); err != nil {
			return err
		}

		return errors.Wrapf(err, "unable to update flow step %s status to discarded", step.ID)
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: inp.FlowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "generating steps for flow",
		},
	}); err != nil {
		l.Error("unable to update status on retry", zap.Error(err))
	}

	l.Debug("generating steps for flow")
	if inp.RetryStep {
		// create new retry step
		// this can be moved into a seprate helper for reusability
		err := c.cloneWorkflowStep(ctx, step, flw)
		if err != nil {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: inp.FlowID,
				Status: app.CompositeStatus{
					Status:                 app.StatusError,
					StatusHumanDescription: "unable to create retry step",
					Metadata: map[string]any{
						"error_message": err.Error(),
					},
				},
			}); err != nil {
				return err
			}
			return errors.Wrapf(err, "unable to create retry step for flow %s", inp.FlowID)
		}
	}

	flowSteps, err := activities.AwaitPkgWorkflowsFlowGetFlowStepsByFlowID(ctx, inp.FlowID)
	if err != nil {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: inp.FlowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "unable to fetch flow step",
				Metadata: map[string]any{
					"error_message": err.Error(),
				},
			},
		}); err != nil {
			return err
		}
		return errors.Errorf("unable to fetch flow steps for flow %s: %v", inp.FlowID, err)
	}

	// get the index of newly created step
	var workflowStartStepNumber int
	for i, step := range flowSteps {
		if step.ID == inp.StepID {
			// if the step was retried it'll start from the new retry step
			// if the step was not retried, it'll start from next step
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
			return errors.Wrap(err, "unable to update status")
		}
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: inp.FlowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "successfully generated all steps",
		},
	}); err != nil {
		return err
	}

	flw.Steps = make([]app.FlowStep, len(flowSteps))
	for i, step := range flowSteps {
		flw.Steps[i] = app.InstallWorkflowStep(step)
	}

	l.Debug("executing steps for flow")
	if err := c.executeFlowSteps(ctx, req, flw, workflowStartStepNumber); err != nil {
		status := app.CompositeStatus{
			Status:                 app.StatusError,
			StatusHumanDescription: "error while executing steps",
			Metadata: map[string]any{
				"error_message": err.Error(),
			},
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID:     inp.FlowID,
			Status: status,
		}); err != nil {
			return err
		}

		return errors.Wrap(err, "unable to execute flow steps")
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: inp.FlowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusSuccess,
			StatusHumanDescription: "successfully executed flow",
		},
	}); err != nil {
		return err
	}

	return nil
}
