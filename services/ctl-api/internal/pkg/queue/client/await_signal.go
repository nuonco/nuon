package client

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

// AwaitSignal is deprecated — callers should use the callback-based await pattern instead.
// This method logs an error with the caller's queue signal ID and returns a non-retryable
// error so the offending call site is surfaced immediately.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) AwaitSignal(ctx context.Context, queueSignalID string) (*handler.FinishedResponse, error) {
	c.l.Error("AwaitSignal is deprecated: use callback-based await instead",
		zap.String("queue_signal_id", queueSignalID),
	)
	return nil, temporal.NewNonRetryableApplicationError(
		fmt.Sprintf("AwaitSignal is deprecated and must not be called (queue_signal_id=%s); use callback-based await instead", queueSignalID),
		"DEPRECATED_AWAIT_SIGNAL", nil)
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) GetQueueSignal(ctx context.Context, id string) (*app.QueueSignal, error) {
	return c.getQueueSignal(ctx, id)
}

func (c *Client) getQueueSignal(ctx context.Context, id string) (*app.QueueSignal, error) {
	var q app.QueueSignal
	if res := c.db.WithContext(ctx).First(&q, "id = ?", id); res.Error != nil {
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get queue signal")
	}

	return &q, nil
}
