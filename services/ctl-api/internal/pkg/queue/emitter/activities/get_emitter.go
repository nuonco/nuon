package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetEmitterRequest struct {
	EmitterID string `validate:"required"`
}

// @temporal-gen activity
// @by-id EmitterID
func (a *Activities) GetEmitter(ctx context.Context, req *GetEmitterRequest) (*app.QueueEmitter, error) {
	var emitter app.QueueEmitter

	if res := a.db.WithContext(ctx).
		Where("id = ?", req.EmitterID).
		First(&emitter); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get emitter")
	}

	return &emitter, nil
}
