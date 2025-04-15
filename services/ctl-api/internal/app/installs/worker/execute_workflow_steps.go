package worker

import (
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

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

	steps, err := activities.AwaitGetInstallWorkflowsStepsByInstallWorkflowID(ctx, sreq.InstallWorkflowID)
	if err != nil {
		return err
	}

	for _, step := range steps {
		if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: sreq.InstallWorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusInProgress,
				StatusHumanDescription: "executing step " + strconv.Itoa(step.Idx),
				Metadata:               map[string]any{},
			},
		}); err != nil {
			return err
		}

		sreq.Signal.InstallWorkflowStepID = step.ID
		if err := w.AwaitExecuteWorkflowStep(ctx, sreq); err != nil {
			if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
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
			}); err != nil {
				return err
			}

			if wkflow.StepErrorBehavior == app.StepErrorBehaviorAbort {
				return err
			}
		}

		if err := statusactivities.AwaitPkgStatusUpdateInstallWorkflowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: sreq.InstallWorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusError,
				StatusHumanDescription: "finished executing step " + strconv.Itoa(step.Idx),
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
