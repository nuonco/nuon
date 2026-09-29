package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type RetryGroupRequest struct {
	InstallWorkflowID string
	StepID            string
}

type RetryGroupResponse struct {
	WorkflowID string `json:"workflow_id"`
	Retryable  bool   `json:"retryable"`
}

func (c *Client) RetryGroup(ctx context.Context, req *RetryGroupRequest) (*RetryGroupResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	var resp RetryGroupResponse
	if err := c.updateWithStartUntilCompleted(ctx, qs, "retry-group", &resp, executeflow.RetryGroupRequest{StepID: req.StepID}); err != nil {
		return nil, fmt.Errorf("unable to get retry-group response: %w", err)
	}

	return &resp, nil
}
