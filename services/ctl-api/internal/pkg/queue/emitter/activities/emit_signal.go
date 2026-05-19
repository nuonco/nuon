package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
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
func (a *Activities) EmitSignal(ctx context.Context, req *EmitSignalRequest) (*EmitSignalResponse, error) {
	// Get the emitter to access its signal template
	var emitter app.QueueEmitter
	if res := a.db.WithContext(ctx).
		Where("id = ?", req.EmitterID).
		First(&emitter); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get emitter")
	}

	if emitter.SignalTemplate.Signal == nil {
		return nil, errors.New("emitter has no signal template configured")
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
		if maxAge := signal.DeriveMaxInFlightAge(emitter.SignalTemplate.Signal); maxAge > 0 {
			stale := make([]*app.QueueSignal, 0, len(existingSignals))
			live := existingSignals[:0]
			now := time.Now()
			for _, s := range existingSignals {
				if now.Sub(s.CreatedAt) > maxAge {
					stale = append(stale, s)
				} else {
					live = append(live, s)
				}
			}
			if len(stale) > 0 {
				staleIDs := make([]string, 0, len(stale))
				for _, s := range stale {
					staleIDs = append(staleIDs, s.ID)
				}
				if res := a.db.WithContext(ctx).Exec(`
					UPDATE queue_signals
					SET status = jsonb_set(status, '{status}', '"error"'::jsonb)
					           || jsonb_build_object('metadata', jsonb_build_object('stale_drop', 'exceeded max_in_flight_age')),
					    deleted_at = extract(epoch from now())::bigint,
					    updated_at = now()
					WHERE id IN (?)`, staleIDs); res.Error != nil {
					return nil, errors.Wrap(res.Error, "unable to mark stale in-flight signals as failed")
				}
				a.l.Warn("dropped stale in-flight signals exceeding max-in-flight age",
					zap.String("emitter-id", req.EmitterID),
					zap.String("queue-id", req.QueueID),
					zap.Int("stale-count", len(stale)),
					zap.Duration("max-in-flight-age", maxAge),
				)
				// Cancel each stale signal's handler workflow so it stops
				// trying to update the now-soft-deleted row.
				for _, s := range stale {
					if err := a.queueClient.CancelHandlerWorkflow(ctx, s.Workflow.Namespace, s.Workflow.ID); err != nil {
						a.l.Warn("unable to cancel handler workflow for stale signal",
							zap.String("queue-signal-id", s.ID),
							zap.String("workflow-id", s.Workflow.ID),
							zap.Error(err),
						)
					}
				}
			}
			existingSignals = live
		}

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
		return nil, errors.Wrap(res.Error, "unable to get queue")
	}

	// Enqueue the signal to the queue using the queue client
	enqueueResp, err := a.queueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID:   req.QueueID,
		Signal:    emitter.SignalTemplate.Signal,
		OwnerID:   queue.OwnerID,
		OwnerType: queue.OwnerType,
	})
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
