package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
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
	// Forward the approval to the step's handler workflow via activity.
	// The step's "approve-plan" handler sets s.approved = true to unblock the step.
	_, err := activities.AwaitForwardApprovePlan(ctx, activities.ForwardApprovePlanRequest{
		StepID:             req.StepID,
		ApprovalResponseID: req.ApprovalResponseID,
		ResponseType:       req.ResponseType,
	})
	if err != nil {
		return nil, err
	}

	return &ApproveStepResponse{WorkflowID: s.InstallWorkflowID}, nil
}
