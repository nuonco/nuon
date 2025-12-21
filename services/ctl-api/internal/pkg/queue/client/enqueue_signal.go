package client

import (
	"context"

	"github.com/pkg/errors"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type EnqueueSignalRequest struct {
	QueueID string        `validate:"required"`
	Signal  signal.Signal `validate:"required"`
}

// @temporal-gen activity
func (c *Client) EnqueueSignal(ctx context.Context, req *EnqueueSignalRequest) (*queue.EnqueueResponse, error) {
	q, err := c.getQueue(ctx, req.QueueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   q.Workflow.ID,
		UpdateName:   queue.EnqueueUpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			req.Signal,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call enqueue handler")
	}

	var resp queue.EnqueueResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable get response")
	}

	return &resp, nil
}
