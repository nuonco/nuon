package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (c *Client) DirectExecuteSignal(ctx context.Context, queueSignalID string) (*queue.DirectExecuteResponse, error) {
	qs, err := c.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	// if already terminal, nothing to do
	if isTerminalStatus(qs.Status.Status) {
		return &queue.DirectExecuteResponse{QueueSignalID: queueSignalID}, nil
	}

	q, err := c.getQueue(ctx, qs.QueueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	rawResp, err := c.tClient.UpdateWithStartWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWithStartWorkflowOptions{
		UpdateOptions: tclient.UpdateWorkflowOptions{
			WorkflowID:   q.Workflow.ID,
			UpdateName:   queue.DirectExecuteUpdateName,
			WaitForStage: tclient.WorkflowUpdateStageCompleted,
			Args: []any{
				queue.DirectExecuteRequest{
					QueueSignalID: queueSignalID,
				},
			},
		},
		StartWorkflowOperation: c.queueStartOperation(q),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call direct execute handler")
	}

	var resp queue.DirectExecuteResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable get response")
	}

	return &resp, nil
}
