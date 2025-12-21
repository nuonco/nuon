package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
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

func AwaitUpdateEmitterStats(ctx workflow.Context, req *UpdateEmitterStatsRequest) (*UpdateEmitterStatsResponse, error) {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Second,
		StartToCloseTimeout:    5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, (&Activities{}).UpdateEmitterStats, req)
	var ret UpdateEmitterStatsResponse
	if err := fut.Get(ctx, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// @temporal-gen activity
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

	// Fetch updated emit count
	var emitter app.QueueEmitter
	if res := a.db.WithContext(ctx).
		Select("emit_count").
		Where("id = ?", req.EmitterID).
		First(&emitter); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get updated emitter")
	}

	a.l.Info("updated emitter stats",
		zap.String("emitter-id", req.EmitterID),
		zap.Int64("emit-count", emitter.EmitCount),
	)

	return &UpdateEmitterStatsResponse{
		EmitCount: emitter.EmitCount,
	}, nil
}
