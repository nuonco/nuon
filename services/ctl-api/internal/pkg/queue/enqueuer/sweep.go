package enqueuer

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// sweep queries for signals that have not been enqueued and processes them.
func (e *Enqueuer) sweep() {
	ctx := context.Background()

	var signals []app.QueueSignal
	if res := e.db.WithContext(ctx).
		Select("id").
		Where("enqueued = ?", false).
		Where("created_at < ?", time.Now().Add(-30*time.Second)).
		Order("created_at ASC").
		Limit(sweepBatchSize).
		Find(&signals); res.Error != nil {
		e.l.Warn("sweep query failed", zap.Error(res.Error))
		return
	}

	if len(signals) > 0 {
		e.l.Warn("sweep found orphaned signals", zap.Int("count", len(signals)))
	}

	for _, s := range signals {
		e.processOne(s.ID)
	}
}
