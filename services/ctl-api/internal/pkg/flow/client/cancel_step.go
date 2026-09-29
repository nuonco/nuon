package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type CancelStepRequest struct {
	InstallWorkflowID string
	StepID            string
}

type CancelStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

func (c *Client) CancelStep(ctx context.Context, req *CancelStepRequest) (*CancelStepResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	var resp CancelStepResponse
	if err := c.updateWithStartUntilCompleted(ctx, qs, "cancel-step", &resp, executeflow.CancelStepRequest{StepID: req.StepID}); err != nil {
		return nil, fmt.Errorf("unable to get cancel-step response: %w", err)
	}

	return &resp, nil
}
