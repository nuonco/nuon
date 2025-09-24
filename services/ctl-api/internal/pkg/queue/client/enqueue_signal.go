package client

import (
	"context"

	"github.com/pkg/errors"

	tclient "go.temporal.io/sdk/client"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
)

func (c *Client) EnqueueSignal(ctx context.Context, queueID string, sig signal.Signal) (*queue.EnqueueResponse, error) {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID: q.Workflow.ID,
		UpdateName: queue.EnqueueUpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			sig,
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
