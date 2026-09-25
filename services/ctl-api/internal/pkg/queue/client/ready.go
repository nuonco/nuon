package client

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/api/serviceerror"

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
			// The queue workflow registers its query handlers after running
			// startup activities, so a query can land before registration.
			// Treat that window the same as "not ready" and keep polling.
			var queryFailed *serviceerror.QueryFailed
			if errors.As(err, &queryFailed) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(500 * time.Millisecond):
				}
				continue
			}
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
