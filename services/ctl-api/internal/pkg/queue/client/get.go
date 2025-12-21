package client

import (
	"context"

	"github.com/pkg/errors"
	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

func (c *Client) GetQueue(ctx context.Context, id string) (*app.Queue, error) {
	q, err := c.getQueue(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	return q, nil
}

func (c *Client) getQueue(ctx context.Context, id string) (*app.Queue, error) {
	var q app.Queue
	if res := c.db.WithContext(ctx).First(&q, "id = ?", id); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue")
	}

	return &q, nil
}

func (c *Client) GetQueueStatus(ctx context.Context, queueID string) (*queue.StatusResponse, error) {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   q.Workflow.ID,
		UpdateName:   queue.StatusHandlerName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			queue.StatusRequest{},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call status handler")
	}

	var resp queue.StatusResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get response")
	}

	return &resp, nil
}
