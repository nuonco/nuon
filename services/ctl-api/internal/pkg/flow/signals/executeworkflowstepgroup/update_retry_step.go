package executeworkflowstepgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// RetryStepRequest is the input for the "retry-step" group update handler.
type RetryStepRequest struct {
	StepID string `json:"step_id"`
}

// RetryStepResponse is the response from the "retry-step" group update handler.
type RetryStepResponse struct {
	Retryable bool   `json:"retryable"`
	Directive string `json:"directive"`
}

// retryStepHandler forwards the retry to the step signal. The step's
// createStepRetryHandler marks the step as retried and writes the terminal
// directive (retry or retry-group). This unblocks the step's Execute(),
// completing the queue signal. The group's sequential loop then reads the
// directive and handles cloning.
//
// This handler does NOT clone — the sequential loop is the single
// authoritative place for clone decisions.
func (s *Signal) retryStepHandler(ctx workflow.Context, req RetryStepRequest) (*RetryStepResponse, error) {
	resp, err := activities.AwaitForwardCreateStepRetry(ctx, activities.ForwardCreateStepRetryRequest{
		StepID: req.StepID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to forward retry to step %s: %w", req.StepID, err)
	}

	return &RetryStepResponse{
		Retryable: true,
		Directive: resp.Directive,
	}, nil
}
