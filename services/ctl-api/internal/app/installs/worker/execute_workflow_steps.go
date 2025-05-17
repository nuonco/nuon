package worker

import (
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"

	"github.com/cockroachdb/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

var WorkflowCancellationErr = fmt.Errorf("workflow cancelled")

func ExecuteWorkflowStepsWorkflowID(req signals.RequestSignal) string {
	return fmt.Sprintf("%s-execute-workflow-steps", req.WorkflowID(req.ID))
}

// @id-callback ExecuteWorkflowStepsWorkflowID
// @temporal-gen workflow
// @execution-timeout 120m
func (w *Workflows) ExecuteWorkflowSteps(ctx workflow.Context, sreq signals.RequestSignal) error {
	wkflow, err := activities.AwaitGetInstallWorkflowByID(ctx, sreq.InstallWorkflowID)
	if err != nil {
		return err
	}

	if wkflow.Status.Status == app.StatusCancelled {
		return WorkflowCancellationErr
	}

	steps, err := activities.AwaitGetInstallWorkflowsStepsByInstallWorkflowID(ctx, sreq.InstallWorkflowID)
	if err != nil {
		return err
	}

	for idx, step := range steps {
		sreq.Signal.WorkflowStepID = step.ID

		ctx = cctx.SetInstallWorkflowWorkflowContext(ctx, &app.InstallWorkflowContext{
			ID:             step.InstallWorkflowID,
			WorkflowStepID: &step.ID,
		})

		if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: sreq.InstallWorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: "executing step " + strconv.Itoa(step.Idx+1),
				Metadata:               map[string]any{},
			},
		}); err != nil {
			return errors.Wrap(err, "unable to update step")
		}

		// handle the ok status, and just mark success + continue
		stepErr := AwaitExecuteWorkflowStep(ctx, sreq)
		if stepErr == nil {
			if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: sreq.InstallWorkflowID,
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
				if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
					ID: cancelStep.ID,
					Status: app.NewCompositeTemporalStatus(ctx, app.StatusNotAttempted, map[string]any{
						"reason": fmt.Sprintf("%s and workflow was configured with abort-on-error.", WorkflowCancellationErr.Error()),
					}),
				}); err != nil {
					return errors.Wrap(err, "unable to cancel step after cancelled workflow")
				}
			}

			if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(cancelCtx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.NewCompositeTemporalStatus(ctx, app.StatusCancelled, map[string]any{
					"reason": WorkflowCancellationErr.Error(),
				}),
			}); err != nil {
				return errors.Wrap(err, "unable to update install workflow step as cancelled")
			}

			return stepErr
		}

		if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusError, map[string]any{
				"reason": "step failed",
			}),
		}); err != nil {
			return errors.Wrap(err, "unable to mark step as error")
		}

		// if the workflow was configured to abort, then go ahead and abort and do not attempt future steps
		if wkflow.StepErrorBehavior == app.StepErrorBehaviorAbort {
			reason := "previous step failed"

			for _, cancelStep := range steps[idx+1:] {
				if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
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
