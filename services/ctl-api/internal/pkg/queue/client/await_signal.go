package client

import (
	"context"

	"github.com/pkg/errors"

	tclient "go.temporal.io/sdk/client"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/handler"
)

func (c *Client) AwaitSignal(ctx context.Context, queueSignalID string) (*handler.FinishedResponse, error) {
	q, err := c.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   q.Workflow.ID,
		UpdateName:   handler.FinishedHandlerName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			handler.FinishedRequest{},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call finished handler")
	}

	var resp handler.FinishedResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable get response")
	}

	return &resp, nil
}

func (c *Client) getQueueSignal(ctx context.Context, id string) (*app.QueueSignal, error) {
	var q app.QueueSignal
	if res := c.db.WithContext(ctx).First(&q, "id = ?", id); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue signal")
	}

	return &q, nil
}
