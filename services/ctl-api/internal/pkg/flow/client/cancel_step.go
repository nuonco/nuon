package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// CancelStepRequest is the input for cancelling a workflow step.
type CancelStepRequest struct {
	InstallWorkflowID string
	StepID            string
}

// CancelStepResponse is the response from the cancel-step update.
type CancelStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// CancelStep sends a "cancel-step" update to the execute-flow handler workflow
// for the given install workflow. The handler workflow forwards the cancellation
// to the step's handler workflow.
func (c *Client) CancelStep(ctx context.Context, req *CancelStepRequest) (*CancelStepResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	handle, err := handler.UpdateWithStart(ctx, c.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "cancel-step",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			executeflow.CancelStepRequest{
				StepID: req.StepID,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to send cancel-step update: %w", err)
	}

	var resp CancelStepResponse
	if err := handle.Get(ctx, &resp); err != nil {
		return nil, fmt.Errorf("unable to get cancel-step response: %w", err)
	}

	return &resp, nil
}
