package client

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ClearQueue cancels all in-flight (non-terminal) signals in the given queue.
// Every affected signal is set to StatusCancelled with a "cancelled by
// clear-queue" description.
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

		// Best-effort Temporal workflow cancel — ignore errors since we
		// unconditionally update the DB status below.
		if _, err := c.CancelSignal(ctx, qs.ID); err != nil {
			c.l.Warn("clear-queue: failed to cancel signal via temporal",
				zap.String("queue_signal_id", qs.ID),
				zap.Error(err))
		}

		// Ensure the DB status is set to cancelled with clear-queue metadata
		// regardless of whether the Temporal cancel succeeded.
		cancelledStatus := app.CompositeStatus{
			CreatedAtTS:            time.Now().Unix(),
			Status:                 app.StatusCancelled,
			StatusHumanDescription: "cancelled by clear-queue",
			Metadata:               map[string]any{"cancelled_by": "clear-queue"},
		}
		if res := c.db.WithContext(ctx).
			Model(&app.QueueSignal{}).
			Where("id = ?", qs.ID).
			Update("status", cancelledStatus); res.Error != nil {
			c.l.Warn("clear-queue: failed to update signal status",
				zap.String("queue_signal_id", qs.ID),
				zap.Error(res.Error))
		}
	}

	return nil
}
