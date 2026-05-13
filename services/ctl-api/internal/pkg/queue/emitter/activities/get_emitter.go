package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetEmitterRequest struct {
	EmitterID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field EmitterID
func (a *Activities) GetEmitter(ctx context.Context, req *GetEmitterRequest) (*app.QueueEmitter, error) {
	var emitter app.QueueEmitter

	if res := a.db.WithContext(ctx).
		Preload("Queue").
		Where("id = ?", req.EmitterID).
		First(&emitter); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get emitter")
	}

	return &emitter, nil
}
