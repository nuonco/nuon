package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

// UnpauseWorkflowRequest is the input for unpausing a workflow.
type UnpauseWorkflowRequest struct {
	InstallWorkflowID string
}

// UnpauseWorkflow sends an "unpause-workflow" update to the execute-flow handler
// workflow. The workflow will resume from the next group.
func (c *Client) UnpauseWorkflow(ctx context.Context, req *UnpauseWorkflowRequest) error {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	if err := c.updateWithStartUntilCompleted(ctx, qs, "unpause-workflow", nil); err != nil {
		return fmt.Errorf("unable to send unpause-workflow update: %w", err)
	}

	return nil
}
