package client

import (
	"context"
	"fmt"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// PauseWorkflowRequest is the input for pausing a workflow.
type PauseWorkflowRequest struct {
	InstallWorkflowID string
}

// PauseWorkflow sends a "pause-workflow" update to the execute-flow handler
// workflow. The workflow will pause after the current group completes.
func (c *Client) PauseWorkflow(ctx context.Context, req *PauseWorkflowRequest) error {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	handle, err := handler.UpdateWithStart(ctx, c.tClient, qs, handler.UpdateWithStartOptions{
		UpdateName:   "pause-workflow",
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return fmt.Errorf("unable to send pause-workflow update: %w", err)
	}

	if err := handle.Get(ctx, nil); err != nil {
		return fmt.Errorf("unable to get pause-workflow response: %w", err)
	}

	return nil
}
