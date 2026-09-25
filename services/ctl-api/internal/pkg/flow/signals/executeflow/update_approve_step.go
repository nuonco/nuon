package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// ApproveStepRequest is the input for the "approve-step" update handler.
type ApproveStepRequest struct {
	StepID             string `json:"step_id"`
	ApprovalResponseID string `json:"approval_response_id"`
	ResponseType       string `json:"response_type"`
}

// ApproveStepResponse is the response from the "approve-step" update handler.
type ApproveStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// approveStepPersistedCancelCheckVersion gates the persisted workflow lookup
// because in-flight histories forwarded approvals without that activity.
const approveStepPersistedCancelCheckVersion = "approve-step-persisted-cancel-check-v1"

func (s *Signal) approveStepHandler(ctx workflow.Context, req ApproveStepRequest) (*ApproveStepResponse, error) {
	defer s.beginUpdate()()

	if s.cancelRequested {
		return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
	}
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if workflow.GetVersion(ctx, approveStepPersistedCancelCheckVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
		// Persisted cancel check: on a run re-entered via update-with-start the
		// in-memory flag is lost, and approving against a cancelled workflow must
		// be rejected before it forwards anywhere.
		flw, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, s.WorkflowID)
		if err != nil {
			return nil, fmt.Errorf("unable to get workflow %s: %w", s.WorkflowID, err)
		}
		if flw.Status.Status == app.StatusCancelled {
			s.cancelRequested = true
			return nil, fmt.Errorf("workflow %s is cancelled", s.WorkflowID)
		}
	}

	// Resident hosts have no live group or step to forward to: the step parked
	// and the response is already persisted, so waking the loop is enough. The
	// scheduler sees the response and the group re-dispatches the step.
	if s.Resident {
		if step.Status.Status != app.AwaitingApproval {
			return nil, fmt.Errorf("step %s is not awaiting approval (status %s)", req.StepID, step.Status.Status)
		}
		if step.Approval == nil || step.Approval.Response == nil {
			return nil, fmt.Errorf("step %s has no approval response to apply", req.StepID)
		}
	} else {
		if _, err := workflowactivities.AwaitForwardApproveStepToGroup(ctx, workflowactivities.ForwardApproveStepToGroupRequest{
			StepID:             req.StepID,
			StepGroupID:        step.WorkflowStepGroupID,
			ApprovalResponseID: req.ApprovalResponseID,
			ResponseType:       req.ResponseType,
		}); err != nil {
			return nil, fmt.Errorf("unable to forward approve-step to group: %w", err)
		}
	}

	// Wake up the parent execute loop to continue after approval.
	s.markResumeRequested(ctx, app.WorkflowRunTypeResume, req.StepID)
	return &ApproveStepResponse{WorkflowID: s.WorkflowID}, nil
}
