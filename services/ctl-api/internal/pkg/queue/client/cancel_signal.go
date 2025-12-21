package client

import (
	"context"

	"github.com/pkg/errors"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

func (c *Client) CancelSignal(ctx context.Context, queueSignalID string) (*handler.CancelResponse, error) {
	q, err := c.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue signal")
	}

	rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   q.Workflow.ID,
		UpdateName:   handler.CancelUpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			&handler.CancelRequest{},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to call cancel handler")
	}

	var resp handler.CancelResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get response")
	}

	return &resp, nil
}
