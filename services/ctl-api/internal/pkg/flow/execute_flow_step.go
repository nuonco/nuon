package flow

import (
	"encoding/json"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
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

	// NOTE(jm): this is for pre-queue workflows.
	// Check SignalJSON length because GORM reads JSONB null as a non-nil *Signal
	// with empty SignalJSON (JSON null != SQL NULL), which would cause
	// json.Unmarshal to fail with "unexpected end of JSON input".
	if step.Signal != nil && len(step.Signal.SignalJSON) > 0 {
		var sig DomainSignal
		if err := json.Unmarshal(step.Signal.SignalJSON, &sig); err != nil {
			return c.handleStepErr(ctx, step.ID, err)
		}

		// TODO(sdboyer) abstract actual dispatch of the signal into here once we can, then remove ExecFn completely
		err := c.ExecFnLegacy(ctx, req, sig, *step)
		if err != nil {
			return c.handleStepErr(ctx, step.ID, errors.Wrapf(err, "error executing step %s", step.Name))
		}
	}

	if step.QueueSignal != nil {
		if c.ExecFn != nil {
			err := c.ExecFn(ctx, step.QueueSignal, *step)
			if err != nil {
				return c.handleStepErr(ctx, step.ID, errors.Wrapf(err, "error executing step %s", step.Name))
			}
		}
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
