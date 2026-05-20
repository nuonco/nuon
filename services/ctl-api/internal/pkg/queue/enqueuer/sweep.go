package enqueuer

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	sweepBatchSize   = 1000
	sweepMaxPages    = 10
	sweepGracePeriod = 30 * time.Second
)

type SweepRequest struct{}

type SweepResponse struct {
	EnqueuedByType map[string]int `json:"enqueued_by_type"`
	TotalEnqueued  int            `json:"total_enqueued"`
	TotalErrors    int            `json:"total_errors"`
	DurationMS     int64          `json:"duration_ms"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) Sweep(ctx context.Context, _ *SweepRequest) (*SweepResponse, error) {
	start := time.Now()
	enqueuedByType := make(map[string]int)
	totalEnqueued := 0
	totalErrors := 0
	cutoff := time.Now().Add(-sweepGracePeriod)

	for page := 0; page < sweepMaxPages; page++ {
		var signals []app.QueueSignal
		if res := a.e.db.WithContext(ctx).
			Select("id", "type").
			Where(map[string]any{"enqueued": false}).
			Where("created_at < ?", cutoff).
			Order("created_at DESC").
			Limit(sweepBatchSize).
			Find(&signals); res.Error != nil {
			return nil, res.Error
		}

		if len(signals) == 0 {
			break
		}

		a.e.l.Info("sweep found orphaned signals",
			zap.Int("count", len(signals)),
			zap.Int("page", page))

		for _, s := range signals {
			if err := a.e.EnqueueInline(ctx, s.ID, EnqueueSourceSweep); err != nil {
				totalErrors++
				a.e.l.Warn("sweep enqueue failed",
					zap.String("queue-signal-id", s.ID),
					zap.Error(err))
			} else {
				totalEnqueued++
				enqueuedByType[string(s.Type)]++
			}
		}

		if len(signals) < sweepBatchSize {
			break
		}
	}

	duration := time.Since(start)

	tags := metrics.ToTags(map[string]string{"general": "true"})
	a.e.mw.Gauge("queue_signals.sweep.enqueued", float64(totalEnqueued), tags)
	a.e.mw.Gauge("queue_signals.sweep.errors", float64(totalErrors), tags)
	a.e.mw.Timing("queue_signals.sweep.duration_ms", duration, tags)

	if totalEnqueued > 0 || totalErrors > 0 {
		a.e.l.Info("sweep completed",
			zap.Int("enqueued", totalEnqueued),
			zap.Int("errors", totalErrors),
			zap.Duration("duration", duration))
	}

	return &SweepResponse{
		EnqueuedByType: enqueuedByType,
		TotalEnqueued:  totalEnqueued,
		TotalErrors:    totalErrors,
		DurationMS:     duration.Milliseconds(),
	}, nil
}
