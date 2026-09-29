package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type CancelGroupRequest struct {
	InstallWorkflowID string
	StepID            string
}

type CancelGroupResponse struct {
	WorkflowID string `json:"workflow_id"`
}

func (c *Client) CancelGroup(ctx context.Context, req *CancelGroupRequest) (*CancelGroupResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	var resp CancelGroupResponse
	if err := c.updateWithStartUntilCompleted(ctx, qs, "cancel-group", &resp, executeflow.CancelGroupRequest{StepID: req.StepID}); err != nil {
		return nil, fmt.Errorf("unable to get cancel-group response: %w", err)
	}

	return &resp, nil
}
