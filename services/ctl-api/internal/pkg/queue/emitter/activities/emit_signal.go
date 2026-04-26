package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
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
		a.l.Info("skipping signal emission - emitter already has in-flight signal",
			zap.String("emitter-id", req.EmitterID),
			zap.String("queue-id", req.QueueID),
			zap.Int("existing-signal-count", len(existingSignals)),
			zap.String("existing-signal-id", existingSignals[0].ID),
		)
		return &EmitSignalResponse{Skipped: true}, nil
	}

	// Enqueue the signal to the queue using the queue client
	enqueueResp, err := a.queueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: req.QueueID,
		Signal:  emitter.SignalTemplate.Signal,
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
