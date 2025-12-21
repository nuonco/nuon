package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateEmitterStatusRequest struct {
	EmitterID string     `validate:"required"`
	Status    app.Status `validate:"required"`
}

type UpdateEmitterStatusResponse struct{}

// @temporal-gen activity
func (a *Activities) UpdateEmitterStatus(ctx context.Context, req *UpdateEmitterStatusRequest) (*UpdateEmitterStatusResponse, error) {
	var emitter app.QueueEmitter
	if res := a.db.WithContext(ctx).First(&emitter, "id = ?", req.EmitterID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get emitter")
	}

	emitter.Status = app.NewCompositeStatus(ctx, req.Status)

	if res := a.db.WithContext(ctx).Save(&emitter); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to update emitter status")
	}

	return &UpdateEmitterStatusResponse{}, nil
}
