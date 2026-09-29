package client

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type PauseWorkflowRequest struct {
	InstallWorkflowID string
}

func (c *Client) PauseWorkflow(ctx context.Context, req *PauseWorkflowRequest) error {
	qs, err := c.findQueueSignalByOwner(ctx, req.InstallWorkflowID, "install_workflows", executeflow.SignalType)
	if err != nil {
		return fmt.Errorf("unable to find execute-flow queue signal: %w", err)
	}

	if err := c.updateWithStartUntilCompleted(ctx, qs, "pause-workflow", nil); err != nil {
		return fmt.Errorf("unable to send pause-workflow update: %w", err)
	}

	return nil
}
