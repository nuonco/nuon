package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

func (c *Client) Unpause(ctx context.Context, queueID string) error {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	update, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   q.Workflow.ID,
		UpdateName:   queue.UnpauseUpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			queue.UnpauseRequest{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to call update handler")
	}

	var resp queue.UnpauseResponse
	if err := update.Get(ctx, &resp); err != nil {
		return errors.Wrap(err, "error waiting for handler to finish")
	}

	return nil
}
