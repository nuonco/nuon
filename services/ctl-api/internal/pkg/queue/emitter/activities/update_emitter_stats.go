package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateEmitterStatsRequest struct {
	EmitterID string `validate:"required"`
}

type UpdateEmitterStatsResponse struct {
	EmitCount int64
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateEmitterStats(ctx context.Context, req *UpdateEmitterStatsRequest) (*UpdateEmitterStatsResponse, error) {
	now := time.Now()

	// Update emit count and last emitted timestamp
	res := a.db.WithContext(ctx).
		Model(&app.QueueEmitter{}).
		Where("id = ?", req.EmitterID).
		Updates(map[string]any{
			"emit_count":      gorm.Expr("emit_count + 1"),
			"last_emitted_at": now,
		})

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to update emitter stats")
	}

	a.l.Info("updated emitter stats",
		zap.String("emitter-id", req.EmitterID),
	)

	return &UpdateEmitterStatsResponse{}, nil
}
