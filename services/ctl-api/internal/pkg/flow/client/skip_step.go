package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type SkipStepRequest struct {
	InstallWorkflowID string
	StepID            string
}

type SkipStepResponse struct {
	WorkflowID string `json:"workflow_id"`
	Skippable  bool   `json:"skippable"`
}

func (c *Client) SkipStep(ctx context.Context, req *SkipStepRequest) (*SkipStepResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	var resp SkipStepResponse
	if err := c.updateWithStartUntilCompleted(ctx, qs, "skip-step", &resp, executeflow.SkipStepRequest{StepID: req.StepID}); err != nil {
		return nil, fmt.Errorf("unable to get skip-step response: %w", err)
	}

	return &resp, nil
}
