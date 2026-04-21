package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CancelStepRequest is the input for the "cancel-step" group update handler.
type CancelStepRequest struct {
	StepID string `json:"step_id"`
}

// CancelStepResponse is the response from the "cancel-step" group update handler.
type CancelStepResponse struct{}

// cancelStepHandler forwards cancellation to the step's handler workflow and
// sets cancelRequested so the group loop stops after the current step.
func (s *Signal) cancelStepHandler(ctx workflow.Context, req CancelStepRequest) (*CancelStepResponse, error) {
	if _, err := workflowactivities.AwaitForwardCancelStep(ctx, workflowactivities.ForwardCancelStepRequest{
		StepID: req.StepID,
	}); err != nil {
		return nil, fmt.Errorf("unable to cancel step %s: %w", req.StepID, err)
	}

	s.cancelRequested = true
	return &CancelStepResponse{}, nil
}
