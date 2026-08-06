package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type EmitSignalRequest struct {
	EmitterID string `validate:"required"`
	QueueID   string `validate:"required"`
}

type EmitSignalResponse struct {
	QueueSignalID string
	WorkflowID    string
	Skipped       bool
}

// @temporal-gen-v2 activity
// @max-retries 10
func (a *Activities) EmitSignal(ctx context.Context, req *EmitSignalRequest) (*EmitSignalResponse, error) {
	// Get the emitter to access its signal template
	var emitter app.QueueEmitter
	if res := a.db.WithContext(ctx).
		Where("id = ?", req.EmitterID).
		First(&emitter); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get emitter")
	}

	if emitter.SignalTemplate.Signal == nil {
		return nil, temporal.NewNonRetryableApplicationError(
			"emitter has no signal template configured",
			"EMITTER_CONFIG_ERROR",
			nil,
		)
	}

	// Check for existing in-flight signals from this emitter to prevent backup
	var existingSignals []*app.QueueSignal
	jdb := generics.NewJSONBQuery(a.db.WithContext(ctx))
	if res := jdb.WhereJSON(generics.JSONBQuery{
		Operator: "IN",
		Field:    "status",
		Path:     "status",
		Value:    []string{string(app.StatusQueued), string(app.StatusInProgress)},
	}).Where(app.QueueSignal{
		EmitterID: &req.EmitterID,
		QueueID:   req.QueueID,
	}).Find(&existingSignals); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to check for existing in-flight signals")
	}

	if len(existingSignals) > 0 {
		staleByReason, live := partitionStaleSignals(
			existingSignals,
			signal.DeriveMaxInFlightAge(emitter.SignalTemplate.Signal),
			time.Now(),
		)

		for reason, ids := range staleByReason {
			// Do NOT soft-delete: a handler may already be running this signal,
			// and removing the row would cause its status-update activities to
			// loop on ErrRecordNotFound. Marking status=error is enough to
			// release EmitSignal's in-flight check; the handler will finish
			// (or its own writes will overwrite this status) naturally.
			if res := a.db.WithContext(ctx).Exec(`
				UPDATE queue_signals
				SET status = jsonb_set(status, '{status}', '"error"'::jsonb)
				           || jsonb_build_object('metadata', jsonb_build_object('stale_drop', ?::text)),
				    updated_at = now()
				WHERE id IN (?)`, reason, ids); res.Error != nil {
				return nil, errors.Wrap(res.Error, "unable to mark stale in-flight signals as failed")
			}
			a.l.Warn("dropped stale in-flight signals",
				zap.String("emitter-id", req.EmitterID),
				zap.String("queue-id", req.QueueID),
				zap.String("reason", reason),
				zap.Int("stale-count", len(ids)),
			)
		}
		existingSignals = live

		if len(existingSignals) > 0 {
			a.l.Info("skipping signal emission - emitter already has in-flight signal",
				zap.String("emitter-id", req.EmitterID),
				zap.String("queue-id", req.QueueID),
				zap.Int("existing-signal-count", len(existingSignals)),
				zap.String("existing-signal-id", existingSignals[0].ID),
			)
			return &EmitSignalResponse{Skipped: true}, nil
		}
	}

	// Look up the queue so we can propagate its owner to the signal.
	var queue app.Queue
	if res := a.db.WithContext(ctx).First(&queue, "id = ?", req.QueueID); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get queue")
	}

	// Enqueue the signal to the queue using the queue client
	enqueueReq := &client.EnqueueSignalRequest{
		QueueID:   req.QueueID,
		Signal:    emitter.SignalTemplate.Signal,
		OwnerID:   queue.OwnerID,
		OwnerType: queue.OwnerType,
		EmitterID: &req.EmitterID,
	}
	if emitter.SignalExpiresIn > 0 {
		expiresAt := time.Now().Add(emitter.SignalExpiresIn)
		enqueueReq.ExpiresAt = &expiresAt
	}
	enqueueResp, err := a.queueClient.EnqueueSignal(ctx, enqueueReq)
	if err != nil {
		return nil, errors.Wrap(err, "unable to enqueue signal to queue")
	}

	a.l.Info("signal emitted to queue",
		zap.String("emitter-id", req.EmitterID),
		zap.String("queue-id", req.QueueID),
		zap.String("queue-signal-id", enqueueResp.ID),
		zap.String("workflow-id", enqueueResp.WorkflowID),
	)

	return &EmitSignalResponse{
		QueueSignalID: enqueueResp.ID,
		WorkflowID:    enqueueResp.WorkflowID,
	}, nil
}

const (
	staleReasonMaxInFlightAge = "exceeded max_in_flight_age"
	staleReasonExpired        = "expired"
)

// partitionStaleSignals splits an emitter's in-flight signals into ones that should
// no longer hold the emitter back, keyed by why, and ones still considered live.
//
// Expiry is otherwise only enforced by the handler itself, so a handler that died
// leaves its row in-flight forever and the emitter never fires again.
func partitionStaleSignals(signals []*app.QueueSignal, maxAge time.Duration, now time.Time) (map[string][]string, []*app.QueueSignal) {
	staleByReason := map[string][]string{}
	live := make([]*app.QueueSignal, 0, len(signals))

	for _, s := range signals {
		switch {
		case maxAge > 0 && now.Sub(s.CreatedAt) > maxAge:
			staleByReason[staleReasonMaxInFlightAge] = append(staleByReason[staleReasonMaxInFlightAge], s.ID)
		case s.ExpiresAt != nil && now.After(*s.ExpiresAt):
			staleByReason[staleReasonExpired] = append(staleByReason[staleReasonExpired], s.ID)
		default:
			live = append(live, s)
		}
	}

	return staleByReason, live
}
