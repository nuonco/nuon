package activities

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateSignalEmitterRequest struct {
	QueueSignalID string `validate:"required"`
	EmitterID     string `validate:"required"`
}

type UpdateSignalEmitterResponse struct {
	Success bool
}

// @temporal-gen activity
func (a *Activities) UpdateSignalEmitter(ctx context.Context, req *UpdateSignalEmitterRequest) (*UpdateSignalEmitterResponse, error) {
	// Update the queue signal to set its emitter relationship
	res := a.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Where("id = ?", req.QueueSignalID).
		Update("emitter_id", req.EmitterID)

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to update queue signal emitter")
	}

	if res.RowsAffected == 0 {
		return nil, errors.New("queue signal not found")
	}

	a.l.Info("updated queue signal with emitter relationship",
		zap.String("queue-signal-id", req.QueueSignalID),
		zap.String("emitter-id", req.EmitterID),
	)

	return &UpdateSignalEmitterResponse{
		Success: true,
	}, nil
}
