package client

import (
	"context"

	"github.com/pkg/errors"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// SleepSignal puts a handler workflow to sleep, causing it to terminate.
func (c *Client) SleepSignal(ctx context.Context, queueSignalID string) error {
	q, err := c.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue signal")
	}

	rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, q.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   q.Workflow.ID,
		UpdateName:   handler.SleepUpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return errors.Wrap(err, "unable to call sleep handler")
	}

	var resp handler.SleepResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return errors.Wrap(err, "unable get response")
	}

	return nil
}
