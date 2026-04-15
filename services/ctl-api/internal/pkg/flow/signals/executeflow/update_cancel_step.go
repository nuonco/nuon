package executeflow

import (
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
	// Forward the cancel to the step's handler workflow via activity.
	_, err := workflowactivities.AwaitForwardCancelStep(ctx, workflowactivities.ForwardCancelStepRequest{
		StepID: req.StepID,
	})
	if err != nil {
		return nil, err
	}

	s.cancelRequested = true
	return &CancelStepResponse{WorkflowID: s.WorkflowID}, nil
}
