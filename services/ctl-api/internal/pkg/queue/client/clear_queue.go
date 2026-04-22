package client

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ClearQueue cancels all in-flight (non-terminal) signals in the given queue.
func (c *Client) ClearQueue(ctx context.Context, queueID string) error {
	var signals []app.QueueSignal
	if res := c.db.WithContext(ctx).
		Where(app.QueueSignal{QueueID: queueID}).
		Find(&signals); res.Error != nil {
		return errors.Wrap(res.Error, "unable to list queue signals")
	}

	for _, qs := range signals {
		if isTerminalStatus(qs.Status.Status) {
			continue
		}

		if _, err := c.CancelSignal(ctx, qs.ID); err != nil {
			c.l.Warn("clear-queue: failed to cancel signal",
				zap.String("queue_signal_id", qs.ID),
				zap.Error(err))
		}
	}

	return nil
}
