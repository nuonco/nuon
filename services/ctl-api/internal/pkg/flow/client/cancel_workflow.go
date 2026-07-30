package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

// CancelWorkflowRequest is the input for cancelling an entire workflow.
type CancelWorkflowRequest struct {
	InstallWorkflowID string
}

// CancelWorkflowResponse is the response from the cancel-workflow update.
type CancelWorkflowResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// CancelWorkflow sends a "cancel-workflow" update to the execute-flow handler
// workflow, cancelling the entire workflow after the current group completes.
func (c *Client) CancelWorkflow(ctx context.Context, req *CancelWorkflowRequest) (*CancelWorkflowResponse, error) {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return nil, fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	var resp CancelWorkflowResponse
	if err := c.updateWithStartUntilCompleted(ctx, qs, "cancel-workflow", &resp); err != nil {
		return nil, fmt.Errorf("unable to get cancel-workflow response: %w", err)
	}

	return &resp, nil
}
