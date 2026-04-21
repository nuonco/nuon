package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// ApproveStepRequest is the input for the "approve-step" group update handler.
type ApproveStepRequest struct {
	StepID             string `json:"step_id"`
	ApprovalResponseID string `json:"approval_response_id"`
	ResponseType       string `json:"response_type"`
}

// ApproveStepResponse is the response from the "approve-step" group update handler.
type ApproveStepResponse struct{}

// approveStepHandler forwards the approval to the step's handler workflow.
// The step will write a directive that the group's sequential loop reads.
func (s *Signal) approveStepHandler(ctx workflow.Context, req ApproveStepRequest) (*ApproveStepResponse, error) {
	if _, err := workflowactivities.AwaitForwardApprovePlan(ctx, workflowactivities.ForwardApprovePlanRequest{
		StepID:             req.StepID,
		ApprovalResponseID: req.ApprovalResponseID,
		ResponseType:       req.ResponseType,
	}); err != nil {
		return nil, fmt.Errorf("unable to approve step %s: %w", req.StepID, err)
	}

	return &ApproveStepResponse{}, nil
}
