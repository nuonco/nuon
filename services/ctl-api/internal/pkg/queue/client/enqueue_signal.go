package client

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
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
	DedupeKey *string

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
	resp, queueSignal, err := c.enqueueSignal(ctx, c.db, req)
	if err != nil {
		return nil, err
	}
	if c.enqueuer != nil && queueSignal != nil {
		c.enqueuer.Send(queueSignal.ID)
	}
	return resp, nil
}

func (c *Client) EnqueueSignalInTransaction(ctx context.Context, tx *gorm.DB, req *EnqueueSignalRequest) (*queue.EnqueueResponse, error) {
	resp, _, err := c.enqueueSignal(ctx, tx, req)
	return resp, err
}

// NotifySignal wakes the queue processor for a signal that was enqueued inside
// a transaction (EnqueueSignalInTransaction cannot notify pre-commit). Call it
// after the transaction commits; without it the signal only runs when the
// enqueuer sweep finds it.
func (c *Client) NotifySignal(queueSignalID string) {
	if c.enqueuer != nil && queueSignalID != "" {
		c.enqueuer.Send(queueSignalID)
	}
}

func (c *Client) enqueueSignal(ctx context.Context, db *gorm.DB, req *EnqueueSignalRequest) (*queue.EnqueueResponse, *app.QueueSignal, error) {
	var q app.Queue
	if err := db.WithContext(ctx).Where(app.Queue{ID: req.QueueID}).First(&q).Error; err != nil {
		return nil, nil, errors.Wrap(err, "unable to get queue")
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
		DedupeKey: req.DedupeKey,
		Status:    status,
		ExpiresAt: req.ExpiresAt,
		Workflow: signaldb.WorkflowRef{
			Namespace:  q.Workflow.Namespace,
			IDTemplate: q.Workflow.ID + "-handler-%s-" + string(req.Signal.Type()) + "-" + hex.EncodeToString(suffix),
		},
		Callback:  req.Callback,
		Callbacks: callbacks,
	}

	create := db.WithContext(ctx)
	if req.IdempotencyKey != "" {
		queueSignal.ID = idempotentQueueSignalID(req.QueueID, req.Signal.Type(), req.IdempotencyKey)
		queueSignal.Workflow.ID = fmt.Sprintf(queueSignal.Workflow.IDTemplate, queueSignal.ID)
		create = create.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		})
	} else if req.DedupeKey != nil && *req.DedupeKey != "" {
		create = create.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "queue_id"}, {Name: "dedupe_key"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{
				clause.Expr{SQL: "deleted_at = 0 AND dedupe_key IS NOT NULL AND dedupe_key <> ''"},
			}},
			DoNothing: true,
		})
	}
	res := create.Create(&queueSignal)
	if res.Error != nil {
		return nil, nil, errors.Wrap(res.Error, "unable to create queue signal")
	}
	if res.RowsAffected == 0 {
		var existing app.QueueSignal
		query := db.WithContext(ctx)
		if req.IdempotencyKey != "" {
			query = query.Unscoped().Where(app.QueueSignal{ID: queueSignal.ID})
		} else {
			query = query.Where(app.QueueSignal{QueueID: req.QueueID, DedupeKey: req.DedupeKey})
		}
		if err := query.First(&existing).Error; err != nil {
			return nil, nil, errors.Wrap(err, "unable to get deduplicated queue signal")
		}
		return &queue.EnqueueResponse{ID: existing.ID, WorkflowID: existing.Workflow.ID, Deduplicated: true}, &existing, nil
	}

	c.mw.Incr("queue.signal.enqueued", metrics.ToTags(map[string]string{
		"signal_type": string(req.Signal.Type()),
		"owner_type":  req.OwnerType,
	}))

	return &queue.EnqueueResponse{
		ID:           queueSignal.ID,
		WorkflowID:   queueSignal.Workflow.ID,
		Deduplicated: false,
	}, &queueSignal, nil
}

func idempotentQueueSignalID(queueID string, signalType signal.SignalType, idempotencyKey string) string {
	sum := sha256.Sum256([]byte(queueID + "\x00" + string(signalType) + "\x00" + idempotencyKey))
	return "qsi" + hex.EncodeToString(sum[:])[:23]
}
