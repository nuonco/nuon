package client

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm/clause"

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
	EmitterID *string

	// IdempotencyKey deduplicates repeated requests for the same signal type and queue.
	IdempotencyKey string `validate:"omitempty,max=255"`

	// Callback describes where the handler should send a Temporal signal on completion.
	// Deprecated: use Callbacks for new code.
	Callback callback.Ref

	// Callbacks supports multiple completion targets.
	Callbacks callback.Refs
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

	// Merge single Callback into Callbacks for backward compat.
	callbacks := req.Callbacks
	if req.Callback.IsSet() {
		found := false
		for _, cb := range callbacks {
			if cb.WorkflowID == req.Callback.WorkflowID && cb.SignalName == req.Callback.SignalName {
				found = true
				break
			}
		}
		if !found {
			callbacks = append(callbacks, req.Callback)
		}
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
		EmitterID: req.EmitterID,
		Status:    status,
		ExpiresAt: req.ExpiresAt,
		Workflow: signaldb.WorkflowRef{
			Namespace:  q.Workflow.Namespace,
			IDTemplate: q.Workflow.ID + "-handler-%s-" + string(req.Signal.Type()) + "-" + hex.EncodeToString(suffix),
		},
		Callback:  req.Callback,
		Callbacks: callbacks,
	}

	db := c.db.WithContext(ctx)
	if req.IdempotencyKey != "" {
		queueSignal.ID = idempotentQueueSignalID(req.QueueID, req.Signal.Type(), req.IdempotencyKey)
		queueSignal.Workflow.ID = fmt.Sprintf(queueSignal.Workflow.IDTemplate, queueSignal.ID)
		db = db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		})
	}

	res := db.Create(&queueSignal)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create queue signal")
	}
	if res.RowsAffected == 0 {
		var existing app.QueueSignal
		if res := c.db.WithContext(ctx).
			Unscoped().
			Where(app.QueueSignal{ID: queueSignal.ID}).
			First(&existing); res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to get idempotent queue signal")
		}
		return &queue.EnqueueResponse{
			ID:           existing.ID,
			WorkflowID:   existing.Workflow.ID,
			Deduplicated: true,
		}, nil
	}

	if c.enqueuer != nil {
		c.enqueuer.Send(queueSignal.ID)
	}

	c.mw.Incr("queue.signal.enqueued", metrics.ToTags(map[string]string{
		"signal_type": string(req.Signal.Type()),
		"owner_type":  req.OwnerType,
	}))

	return &queue.EnqueueResponse{
		ID:           queueSignal.ID,
		WorkflowID:   queueSignal.Workflow.ID,
		Deduplicated: false,
	}, nil
}

func idempotentQueueSignalID(queueID string, signalType signal.SignalType, idempotencyKey string) string {
	sum := sha256.Sum256([]byte(queueID + "\x00" + string(signalType) + "\x00" + idempotencyKey))
	return "qsi" + hex.EncodeToString(sum[:])[:23]
}
