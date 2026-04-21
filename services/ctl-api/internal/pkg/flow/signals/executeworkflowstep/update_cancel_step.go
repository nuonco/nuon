package executeworkflowstep

import (
	"go.temporal.io/sdk/workflow"
)

// CancelStepRequest is the input for the "cancel-step" update handler.
type CancelStepRequest struct{}

// CancelStepResponse is the response from the "cancel-step" update handler.
type CancelStepResponse struct{}

// cancelStepHandler delegates to the signal's Cancel() method which propagates
// cancellation to the inner (target) signal and updates step status.
func (s *Signal) cancelStepHandler(ctx workflow.Context, req CancelStepRequest) (*CancelStepResponse, error) {
	s.Cancel(ctx)
	return &CancelStepResponse{}, nil
}
