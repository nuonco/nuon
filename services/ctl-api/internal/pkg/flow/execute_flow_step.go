package flow

import (
	"encoding/json"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (c *WorkflowConductor[DomainSignal]) executeStep(ctx workflow.Context, req eventloop.EventLoopRequest, step *app.WorkflowStep) error {
	defer func() {
		c.checkStepCancellation(ctx, step.ID)
	}()

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepStartedAtByID(ctx, step.ID); err != nil {
		return err
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status: app.StatusInProgress,
		},
	}); err != nil {
		return err
	}

	if step.ExecutionType == app.WorkflowStepExecutionTypeSkipped {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status: app.StatusSuccess,
			},
		}); err != nil {
			return err
		}
		return nil
	}

	var sig DomainSignal
	if err := json.Unmarshal(step.Signal.SignalJSON, &sig); err != nil {
		return c.handleStepErr(ctx, step.ID, err)
	}

	// TODO(sdboyer) abstract actual dispatch of the signal into here once we can, then remove ExecFn completely
	err := c.ExecFn(ctx, req, sig, *step)
	if err != nil {
		return c.handleStepErr(ctx, step.ID, errors.Wrapf(err, "error executing step %s", step.Name))
	}

	return nil
}

func (c *WorkflowConductor[DomainSignal]) handleStepErr(ctx workflow.Context, stepID string, err error) error {
	if statusErr := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: stepID,
		Status: app.CompositeStatus{
			Status: app.StatusError,
			Metadata: map[string]any{
				"err_message": err.Error(),
			},
		},
	}); statusErr != nil {
		return status.WrapStatusErr(err, statusErr)
	}

	return err
}
