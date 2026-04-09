package client

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) QueueReady(ctx context.Context, queueID string) error {
	q, err := c.getQueue(ctx, queueID)
	if err != nil {
		return errors.Wrap(err, "unable to get queue")
	}

	for {
		resp, err := c.tClient.QueryWorkflowInNamespace(ctx, q.Workflow.Namespace, q.Workflow.ID, "", queue.ReadyHandlerName)
		if err != nil {
			return errors.Wrap(err, "unable to query ready handler")
		}

		var readyResp queue.ReadyResponse
		if err := resp.Get(&readyResp); err != nil {
			return errors.Wrap(err, "unable to decode ready response")
		}

		if readyResp.Ready {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}
