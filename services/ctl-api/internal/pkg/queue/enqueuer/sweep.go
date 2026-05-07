package enqueuer

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const sweepBatchSize = 100

type SweepRequest struct{}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) Sweep(ctx context.Context, _ *SweepRequest) error {
	var signals []app.QueueSignal
	if res := a.e.db.WithContext(ctx).
		Select("id").
		Where("enqueued = ?", false).
		Where("created_at < ?", time.Now().Add(-30*time.Second)).
		Order("created_at ASC").
		Limit(sweepBatchSize).
		Find(&signals); res.Error != nil {
		return res.Error
	}

	if len(signals) == 0 {
		return nil
	}

	a.e.l.Info("sweep found orphaned signals", zap.Int("count", len(signals)))

	for _, s := range signals {
		if err := a.e.EnqueueInline(ctx, s.ID, EnqueueSourceSweep); err != nil {
			a.e.l.Warn("sweep enqueue failed",
				zap.String("queue-signal-id", s.ID),
				zap.Error(err))
		}
	}

	return nil
}
