package executeworkflowstepgroup

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CancelStepRequest is the input for the "cancel-step" group update handler.
type CancelStepRequest struct {
	StepID string `json:"step_id"`
}

// CancelStepResponse is the response from the "cancel-step" group update handler.
type CancelStepResponse struct{}

// cancelStepHandler sets cancelRequested and userActionReceived immediately,
// then propagates cancellation to the step handler asynchronously.
// Setting userActionReceived wakes awaitUserAction if the group is blocked
// waiting for user input after a step failure.
func (s *Signal) cancelStepHandler(ctx workflow.Context, req CancelStepRequest) (*CancelStepResponse, error) {
	s.cancelRequested = true

	l, _ := log.WorkflowLogger(ctx)

	if _, err := workflowactivities.AwaitForwardCancelStep(ctx, workflowactivities.ForwardCancelStepRequest{
		StepID: req.StepID,
	}); err != nil {
		if l != nil {
			l.Warn("cancel-step: unable to forward cancel to step",
				zap.String("step_id", req.StepID),
				zap.Error(err))
		}
	}

	return &CancelStepResponse{}, nil
}
