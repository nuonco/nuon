package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) QueueReady(ctx context.Context, queueID string) error {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	rawResp, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.ReadyHandlerName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				queue.ReadyRequest{},
			},
		},
		StartWorkflowOperation: c.queueStartOperation(q),
	})
	if err != nil {
		return errors.Wrap(err, "unable to call query handler")
	}

	var resp queue.ReadyResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return errors.Wrap(err, "unable get response")
	}

	return nil
}
