package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// Safety-net TTL: bounds how long a stranded claim (e.g. workflow terminated
// before the after-phase hook fires) can pin an emitter.
const InFlightClaimTTL = 1 * time.Hour

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
	var emitter app.QueueEmitter
	if res := a.db.WithContext(ctx).
		Where("id = ?", req.EmitterID).
		First(&emitter); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get emitter")
	}

	if emitter.SignalTemplate.Signal == nil {
		return nil, errors.New("emitter has no signal template configured")
	}

	claim := a.db.WithContext(ctx).
		Model(&app.QueueEmitter{}).
		Where("id = ?", req.EmitterID).
		Where("in_flight_claimed_at IS NULL OR in_flight_claimed_at < ?", time.Now().Add(-InFlightClaimTTL)).
		Update("in_flight_claimed_at", time.Now())
	if claim.Error != nil {
		return nil, errors.Wrap(claim.Error, "unable to claim in-flight slot")
	}
	if claim.RowsAffected == 0 {
		a.l.Info("skipping signal emission - emitter already has in-flight signal",
			zap.String("emitter-id", req.EmitterID),
			zap.String("queue-id", req.QueueID),
		)
		return &EmitSignalResponse{Skipped: true}, nil
	}

	var queue app.Queue
	if res := a.db.WithContext(ctx).First(&queue, "id = ?", req.QueueID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue")
	}

	enqueueResp, err := a.queueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID:   req.QueueID,
		Signal:    emitter.SignalTemplate.Signal,
		OwnerID:   queue.OwnerID,
		OwnerType: queue.OwnerType,
	})
	if err != nil {
		_ = a.db.WithContext(ctx).
			Model(&app.QueueEmitter{}).
			Where("id = ?", req.EmitterID).
			Update("in_flight_claimed_at", nil).Error
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
