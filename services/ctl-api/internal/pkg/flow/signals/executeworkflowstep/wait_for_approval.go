package executeworkflowstep

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// waitForApprovalResponse waits for an approval response reactively using the
// "approve-plan" update handler, or falls back to the polling child workflow
// for non-queue paths.
func (s *Signal) waitForApprovalResponse(ctx workflow.Context, flw *app.Workflow, step *app.WorkflowStep) (*app.WorkflowStepApprovalResponse, error) {
	// Handle auto-approve
	if flw.ApprovalOption == app.InstallApprovalOptionApproveAll {
		return s.handleAutoApproval(ctx, flw, step)
	}

	// Wait for approve-plan update handler to set s.approved (30-day deadline)
	ok, err := workflow.AwaitWithTimeout(ctx, 30*24*time.Hour, func() bool {
		return s.approved
	})
	if err != nil {
		return nil, fmt.Errorf("error waiting for approval for step %s: %w", step.ID, err)
	}
	if !ok {
		statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.NewCompositeTemporalStatus(ctx, app.WorkflowStepApprovalStatusApprovalExpired, map[string]any{
				"err_message": "approval was not accepted",
			}),
		})
		return nil, fmt.Errorf("approval timed out for step %s", step.ID)
	}

	// Fetch the step from DB to get the approval response.
	// The approvalResponseID hint confirms the response exists.
	step, err = activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, step.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get step after approval")
	}
	if step.Approval == nil || step.Approval.Response == nil {
		return nil, errors.New("approval response not found after update")
	}
	return step.Approval.Response, nil
}

// handleAutoApproval handles auto-approve for workflows with ApproveAll option.
func (s *Signal) handleAutoApproval(ctx workflow.Context, flw *app.Workflow, step *app.WorkflowStep) (*app.WorkflowStepApprovalResponse, error) {
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: flw.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusApproved,
			StatusHumanDescription: "auto approved for step " + strconv.Itoa(step.Idx+1),
			Metadata: map[string]any{
				"step_idx": step.Idx,
				"status":   "auto-approved",
			},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "unable to update flow status for auto-approval")
	}

	resp, err := activities.AwaitCreateApprovalResponse(ctx, activities.CreateStepApprovalResponseRequest{
		StepApprovalID: step.Approval.ID,
		Type:           app.WorkflowStepApprovalResponseTypeApprove,
		Note:           "auto-approved",
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create auto-approval response")
	}
	return resp, nil
}
