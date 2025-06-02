package flow

import (
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"

	"github.com/cockroachdb/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/flow/activities"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

var FlowCancellationErr = fmt.Errorf("flow cancelled")

func (c *FlowConductor[DomainSignal]) executeSteps(ctx workflow.Context, req eventloop.EventLoopRequest, flw *app.Flow) error {
	if flw.Status.Status == app.StatusCancelled {
		return FlowCancellationErr
	}

	steps, err := activities.AwaitGetFlowStepsByFlowID(ctx, flw.ID)
	if err != nil {
		return err
	}

	for idx, step := range steps {
		ctx = cctx.SetFlowContextWithinWorkflow(ctx, &app.FlowContext{
			// ID:         step.FlowID, // TODO(sdboyer) REFACTOR
			ID:         step.InstallWorkflowID,
			FlowStepID: &step.ID,
		})

		if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: flw.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: "executing step " + strconv.Itoa(step.Idx+1),
				Metadata:               map[string]any{},
			},
		}); err != nil {
			return errors.Wrap(err, "unable to update step")
		}

		// handle the ok status, and just mark success + continue
		stepErr := c.executeStep(ctx, req, &step)
		if stepErr == nil {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: flw.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusSuccess,
					StatusHumanDescription: "finished executing step " + strconv.Itoa(step.Idx+1),
					Metadata: map[string]any{
						"step_idx": step.Idx,
						"status":   "ok",
					},
				},
			}); err != nil {
				return errors.Wrap(err, "unable to update step to success status")
			}

			continue
		}

		// handle the cancellation status
		if errors.Is(ctx.Err(), workflow.ErrCanceled) {
			cancelCtx, _ := workflow.NewDisconnectedContext(ctx)

			// cancel all steps
			for _, cancelStep := range steps[idx+1:] {
				if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
					ID: cancelStep.ID,
					Status: app.NewCompositeTemporalStatus(ctx, app.StatusNotAttempted, map[string]any{
						"reason": fmt.Sprintf("%s and workflow was configured with abort-on-error.", FlowCancellationErr.Error()),
					}),
				}); err != nil {
					return errors.Wrap(err, "unable to cancel step after cancelled workflow")
				}
			}

			// update the workflow status
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.NewCompositeTemporalStatus(ctx, app.StatusCancelled, map[string]any{
					"reason": FlowCancellationErr.Error(),
				}),
			}); err != nil {
				return errors.Wrap(err, "unable to update install workflow step as cancelled")
			}

			// close the workflow
			if err := activities.AwaitUpdateFlowFinishedAtByID(ctx, step.InstallWorkflowID); err != nil {
				return errors.Wrap(err, "unable to set finished at on the workflow")
			}
			return stepErr
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusError, map[string]any{
				"reason": "step failed",
			}),
		}); err != nil {
			return errors.Wrap(err, "unable to mark step as error")
		}

		// if the workflow was configured to abort, then go ahead and abort and do not attempt future steps
		if flw.StepErrorBehavior == app.StepErrorBehaviorAbort {
			reason := "previous step failed"

			for _, cancelStep := range steps[idx+1:] {
				if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: cancelStep.ID,
					Status: app.NewCompositeTemporalStatus(ctx, app.StatusNotAttempted, map[string]any{
						"reason": fmt.Sprintf("%s and workflow was configured with abort-on-error.", reason),
					}),
				}); err != nil {
					return errors.Wrap(err, "unable to update status")
				}
			}
			return stepErr
		}

		continue
	}

	return nil
}
