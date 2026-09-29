package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type ApproveStepRequest struct {
	StepID             string `json:"step_id"`
	ApprovalResponseID string `json:"approval_response_id"`
	ResponseType       string `json:"response_type"`
}

type ApproveStepResponse struct{}

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
