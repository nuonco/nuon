package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type MarkEmitterFiredRequest struct {
	EmitterID string `validate:"required"`
}

type MarkEmitterFiredResponse struct {
	Success bool
}

// @temporal-gen-v2 activity
func (a *Activities) MarkEmitterFired(ctx context.Context, req *MarkEmitterFiredRequest) (*MarkEmitterFiredResponse, error) {
	now := time.Now()

	res := a.db.WithContext(ctx).
		Model(&app.QueueEmitter{}).
		Where("id = ?", req.EmitterID).
		Updates(map[string]any{
			"fired":           true,
			"last_emitted_at": now,
			"status":          app.NewCompositeStatus(ctx, app.StatusSuccess),
		})

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to mark emitter as fired")
	}

	if res.RowsAffected == 0 {
		return nil, generics.TemporalDoNotRetry(fmt.Errorf("emitter %s not found", req.EmitterID))
	}

	a.l.Info("marked emitter as fired",
		zap.String("emitter-id", req.EmitterID),
	)

	return &MarkEmitterFiredResponse{
		Success: true,
	}, nil
}
