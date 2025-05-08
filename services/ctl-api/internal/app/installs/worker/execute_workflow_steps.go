package worker

import (
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/cockroachdb/errors"
	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

func ExecuteWorkflowStepsWorkflowID(req signals.RequestSignal) string {
	return fmt.Sprintf("%s-execute-workflow-steps", req.WorkflowID(req.ID))
}

// @id-callback ExecuteWorkflowStepsWorkflowID
// @temporal-gen workflow
// @execution-timeout 120m
func (w *Workflows) ExecuteWorkflowSteps(ctx workflow.Context, sreq signals.RequestSignal) error {
	l := temporalzap.GetWorkflowLogger(ctx)
	wkflow, err := activities.AwaitGetInstallWorkflowByID(ctx, sreq.InstallWorkflowID)
	if err != nil {
		return err
	}

	steps, err := activities.AwaitGetInstallWorkflowsStepsByInstallWorkflowID(ctx, sreq.InstallWorkflowID)
	if err != nil {
		return err
	}

	for idx, step := range steps {
		var cancel, steperr error
		sreq.Signal.WorkflowStepID = step.ID

		selector := workflow.NewSelector(ctx)
		fut, sel := workflow.NewFuture(ctx)

		workflow.Go(ctx, func(ctx workflow.Context) {
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
				sel.SetError(err)
				return
			}

			if err := w.AwaitExecuteWorkflowStep(ctx, sreq); err != nil {
				if uerr := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: sreq.InstallWorkflowID,
					Status: app.CompositeStatus{
						Status:                 app.StatusError,
						StatusHumanDescription: "error executing step " + strconv.Itoa(step.Idx),
						Metadata: map[string]any{
							"step_idx":      step.Idx,
							"status":        "error",
							"error_message": err.Error(),
						},
					},
				}); uerr == nil {
					sel.SetError(err)
				} else {
					sel.SetError(errors.Join(err, uerr))
				}
				return
			}

			sel.Set(nil, nil)
		})

		// Two branches on the select - one waits for the step to finish by way of the future,
		// and the other waits for a cancellation signal.
		selector.AddFuture(fut, func(f workflow.Future) {
			steperr = f.Get(ctx, nil)
		})
		selector.AddReceive(ctx.Done(), func(_ workflow.ReceiveChannel, _ bool) {
			cancel = ctx.Err()
		})
		selector.Select(ctx)

		if cancel != nil || steperr != nil {
			var reason string
			if cancel != nil {
				reason = "workflow cancellation was requested"
				steperr = errors.New(reason)

				// Necessary in order to act after a cancellation
				ctx, _ = workflow.NewDisconnectedContext(ctx)
				statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: step.ID,
					Status: app.NewCompositeTemporalStatus(ctx, app.StatusCancelled, map[string]any{
						"reason": reason,
					}),
				})
			} else {
				reason = "previous step failed"
				l.Error("error executing step", zap.String("step", step.ID), zap.Error(err))
			}
			// cancel all steps
			if wkflow.StepErrorBehavior == app.StepErrorBehaviorAbort {
				for _, cancelStep := range steps[idx+1:] {
					statusactivities.AwaitPkgStatusUpdateInstallWorkflowStepStatus(ctx, statusactivities.UpdateStatusRequest{
						ID: cancelStep.ID,
						Status: app.NewCompositeTemporalStatus(ctx, app.StatusNotAttempted, map[string]any{
							"reason": fmt.Sprintf("%s and workflow was configured with abort-on-error.", reason),
						}),
					})
				}
			}
			return steperr
		}

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
			return err
		}
	}

	return nil
}
