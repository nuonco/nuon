package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstep"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// IsRetryableRequest is the input for checking step retryability.
type IsRetryableRequest struct {
	StepID string
}

// IsRetryableResponse is the response indicating step retryability.
type IsRetryableResponse struct {
	Retryable bool   `json:"retryable"`
	Skippable bool   `json:"skippable"`
	StepID    string `json:"step_id"`
}

// IsRetryable sends an "is-retryable" update to the execute-workflow-step
// handler workflow to check if the step can be retried.
func (c *Client) IsRetryable(ctx context.Context, req *IsRetryableRequest) (*IsRetryableResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.StepID, "install_workflow_steps", executeworkflowstep.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find step queue signal: %w", err)
	}

	handle, err := handler.UpdateWithStart(ctx, c.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "is-retryable",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send is-retryable update: %w", err)
	}

	var stepResp executeworkflowstep.IsRetryableResponse
	if err := handle.Get(ctx, &stepResp); err != nil {
		return nil, fmt.Errorf("unable to get is-retryable response: %w", err)
	}

	return &IsRetryableResponse{
		Retryable: stepResp.Retryable,
		Skippable: stepResp.Skippable,
		StepID:    stepResp.StepID,
	}, nil
}
