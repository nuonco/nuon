package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

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

	handle, err := c.tClient.UpdateWorkflowInNamespace(ctx, qs.Workflow.Namespace,
		tclient.UpdateWorkflowOptions{
			WorkflowID:   qs.Workflow.ID,
			UpdateName:   "unpause-workflow",
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
		})
	if err != nil {
		return fmt.Errorf("unable to send unpause-workflow update: %w", err)
	}

	if err := handle.Get(ctx, nil); err != nil {
		return fmt.Errorf("unable to get unpause-workflow response: %w", err)
	}

	return nil
}
