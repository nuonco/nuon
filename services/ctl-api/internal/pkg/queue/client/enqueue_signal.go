package client

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuecctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type EnqueueSignalRequest struct {
	QueueID   string        `validate:"required"`
	Signal    signal.Signal `validate:"required"`
	OwnerID   string
	OwnerType string
	ExpiresAt *time.Time

	// Callback describes where the handler should send a Temporal signal on completion.
	Callback callback.Ref
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) EnqueueSignal(ctx context.Context, req *EnqueueSignalRequest) (*queue.EnqueueResponse, error) {
	q, err := c.getQueue(ctx, req.QueueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	// Create the QueueSignal record in the DB directly so we can return the
	// signal ID without waiting for the queue workflow to process it.
	suffix := make([]byte, 3)
	_, _ = rand.Read(suffix)

	status := app.NewCompositeStatus(ctx, app.StatusQueued)
	if t, ok := req.Signal.(signal.SignalWithTimeout); ok {
		if status.Metadata == nil {
			status.Metadata = make(map[string]any)
		}
		status.Metadata["timeout_ns"] = t.Timeout().Nanoseconds()
	}

	queueSignal := app.QueueSignal{
		SignalContext: queuecctx.FromContext(ctx),
		Signal: signaldb.SignalData{
			Signal: req.Signal,
		},
		QueueID:   req.QueueID,
		Type:      req.Signal.Type(),
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
		Status:    status,
		ExpiresAt: req.ExpiresAt,
		Workflow: signaldb.WorkflowRef{
			Namespace:  q.Workflow.Namespace,
			IDTemplate: q.Workflow.ID + "-handler-%s-" + string(req.Signal.Type()) + "-" + hex.EncodeToString(suffix),
		},
		Callback: req.Callback,
	}

	if res := c.db.WithContext(ctx).Create(&queueSignal); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create queue signal")
	}

	if c.enqueuer != nil {
		c.enqueuer.Send(queueSignal.ID)
	}

	c.mw.Incr("queue.signal.enqueued", metrics.ToTags(map[string]string{
		"signal_type": string(req.Signal.Type()),
		"owner_type":  req.OwnerType,
	}))

	return &queue.EnqueueResponse{
		ID:         queueSignal.ID,
		WorkflowID: queueSignal.Workflow.ID,
	}, nil
}
