package flow

import (
	"fmt"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	activities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/flow/activities"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

var NotApprovedErr error = fmt.Errorf("Not approved")

// executeFlowStep executes a single step in the flow. It handles the execution of the step, updates the status, and waits for approval if necessary.
// It returns true if the step needs to be retried (in case of approval steps), false otherwise.
func (c *FlowConductor[DomainSignal]) executeFlowStep(ctx workflow.Context, req eventloop.EventLoopRequest, idx int, step *app.InstallWorkflowStep, flw *app.Flow) (bool, error) {
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: "executing step " + strconv.Itoa(step.Idx+1),
			Metadata:               map[string]any{},
		},
	}); err != nil {
		return false, errors.Wrap(err, "unable to update step")
	}

	// handle the ok status, and just mark success + continue
	stepErr := c.executeStep(ctx, req, step)
	if stepErr != nil {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.StatusError, map[string]any{
				"reason": "step failed",
			}),
		}); err != nil {
			return false, errors.Wrap(err, "unable to mark step as error")
		}
	}

	if step.ExecutionType != app.FlowStepExecutionTypeApproval {
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
			return false, errors.Wrap(err, "unable to update step to success status")
		}

		return false, nil
	}

	// update the status to awaiting
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusAwaitingResponse,
			StatusHumanDescription: "awaiting approval " + strconv.Itoa(step.Idx+1),
			Metadata: map[string]any{
				"step_idx": step.Idx,
				"status":   "ok",
			},
		},
	}); err != nil {
		return false, errors.Wrap(err, "unable to update step to success status")
	}
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusAwaitingResponse,
			StatusHumanDescription: "awaiting approval " + strconv.Itoa(step.Idx+1),
			Metadata: map[string]any{
				"step_idx": step.Idx,
				"status":   "ok",
			},
		},
	}); err != nil {
		return false, errors.Wrap(err, "unable to update step to success status")
	}

	resp, err := c.waitForApprovalResponse(ctx, flw, step, idx)
	if err != nil {
		return false, err
	}

	if resp.Type == app.InstallWorkflowStepApprovalResponseTypeApprove {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.WorkflowStepApprovalStatusApproved,
				StatusHumanDescription: "approved " + strconv.Itoa(step.Idx+1),
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "ok",
				},
			},
		}); err != nil {
			return false, errors.Wrap(err, "unable to update step to success status")
		}
		return false, nil
	}

	// aproval response retry flow
	if resp.Type == app.InstallWorkflowStepApprovalResponseTypeRetryPlan {
		// cloned step which will be retried next
		err := c.cloneWorkflowStep(ctx, step, flw)
		if err != nil {
			return false, errors.Wrap(err, "unable to clone step for retry plan")
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.WorkflowStepApprovalStatusApprovalRetryPlan,
				StatusHumanDescription: "retrying " + strconv.Itoa(step.Idx),
				Metadata: map[string]any{
					"step_idx": step.Idx,
					"status":   "retrying",
				},
			},
		}); err != nil {
			return false, errors.Wrap(err, "unable to update step to retry plan status")
		}

		if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
			StepID:            step.ID,
			Status:            app.InstallDeployStatusV2Noop,
			StatusDescription: "Retrying step " + strconv.Itoa(step.Idx),
		}); err != nil {
			return false, errors.Wrap(err, "unable to update step target status")
		}

		return true, nil
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalDenied, map[string]any{
			"reason": "approval denied",
		}),
	}); err != nil {
		return false, errors.Wrap(err, "unable to update")
	}

	return false, NotApprovedErr
}

func (c *FlowConductor[DomainSignal]) cloneWorkflowStep(ctx workflow.Context, step *app.InstallWorkflowStep, flw *app.Flow) error {
	_, err := activities.AwaitPkgWorkflowsFlowCreateFlowStep(ctx, activities.CreateFlowStepRequest{
		FlowID:        flw.ID,
		OwnerID:       flw.OwnerID,
		OwnerType:     flw.OwnerType,
		Name:          fmt.Sprintf("%s (retry)-%d", step.Name, time.Now().Unix()),
		Signal:        step.Signal,
		Status:        step.Status,
		Idx:           step.Idx,
		ExecutionType: step.ExecutionType,
		Metadata:      step.Metadata,
	})
	return err
}
