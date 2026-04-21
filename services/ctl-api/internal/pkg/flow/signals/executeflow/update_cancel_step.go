package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CancelStepRequest is the input for the "cancel-step" update handler.
type CancelStepRequest struct {
	StepID string `json:"step_id"`
}

// CancelStepResponse is the response from the "cancel-step" update handler.
type CancelStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

func (s *Signal) cancelStepHandler(ctx workflow.Context, req CancelStepRequest) (*CancelStepResponse, error) {
	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		return nil, fmt.Errorf("unable to get step %s: %w", req.StepID, err)
	}

	_, err = workflowactivities.AwaitForwardCancelStepToGroup(ctx, workflowactivities.ForwardCancelStepToGroupRequest{
		StepID:      req.StepID,
		StepGroupID: step.WorkflowStepGroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to forward cancel-step to group: %w", err)
	}

	s.cancelRequested = true
	return &CancelStepResponse{WorkflowID: s.WorkflowID}, nil
}
