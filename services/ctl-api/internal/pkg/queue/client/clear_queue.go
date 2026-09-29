package client

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (c *Client) ClearQueue(ctx context.Context, queueID string) (int, error) {
	var signals []app.QueueSignal
	if res := c.db.WithContext(ctx).
		Where(app.QueueSignal{QueueID: queueID}).
		Find(&signals); res.Error != nil {
		return 0, errors.Wrap(res.Error, "unable to list queue signals")
	}

	cancelled := 0
	for _, qs := range signals {
		if isTerminalStatus(qs.Status.Status) {
			continue
		}

		if _, err := c.CancelSignal(ctx, qs.ID); err != nil {
			c.l.Warn("clear-queue: failed to cancel signal via temporal",
				zap.String("queue_signal_id", qs.ID),
				zap.Error(err))
		}

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

		cancelled++
	}

	return cancelled, nil
}
