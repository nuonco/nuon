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

func (s *Signal) approveStepHandler(ctx workflow.Context, req ApproveStepRequest) (*ApproveStepResponse, error) {
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	if _, err := workflowactivities.AwaitForwardApproveStepToGroup(ctx, workflowactivities.ForwardApproveStepToGroupRequest{
		StepID:             req.StepID,
		StepGroupID:        step.WorkflowStepGroupID,
		ApprovalResponseID: req.ApprovalResponseID,
		ResponseType:       req.ResponseType,
	}); err != nil {
		return nil, fmt.Errorf("unable to forward approve-step to group: %w", err)
	}

	// Wake up the parent execute loop to continue after approval.
	s.resumeRequested = true
	s.resumeRunType = app.WorkflowRunTypeResume
	s.resumeStepID = req.StepID
	s.resumeStartIdx = s.findGroupPositionForStep(ctx, req.StepID)
	return &ApproveStepResponse{WorkflowID: s.WorkflowID}, nil
}
