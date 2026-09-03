package enqueuer

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const fullSweepBatchSize = 5000

type FullSweepRequest struct{}

type FullSweepResponse struct {
	Enqueued   int   `json:"enqueued"`
	Errors     int   `json:"errors"`
	DurationMS int64 `json:"duration_ms"`
}

// FullSweep processes all unenqueued signals with no grace period or page cap.
// It is intended to be called from the admin dashboard to recover from large backlogs.
func (e *Enqueuer) FullSweep(ctx context.Context) (*FullSweepResponse, error) {
	start := time.Now()
	totalEnqueued := 0
	totalErrors := 0

	for {
		var signals []app.QueueSignal
		if res := e.db.WithContext(ctx).
			Select("id").
			Where(map[string]any{"enqueued": false}).
			Order("created_at DESC").
			Limit(fullSweepBatchSize).
			Find(&signals); res.Error != nil {
			return nil, res.Error
		}

		if len(signals) == 0 {
			break
		}

		e.l.Info("full sweep processing batch",
			zap.Int("count", len(signals)),
			zap.Int("total_enqueued_so_far", totalEnqueued))

		for _, s := range signals {
			if err := e.EnqueueInline(ctx, s.ID, EnqueueSourceSweep); err != nil {
				totalErrors++
				e.l.Warn("full sweep enqueue failed",
					zap.String("queue-signal-id", s.ID),
					zap.Error(err))
			} else {
				totalEnqueued++
			}
		}

		if len(signals) < fullSweepBatchSize {
			break
		}
	}

	duration := time.Since(start)

	tags := metrics.ToTags(map[string]string{"general": "true"})
	e.mw.Gauge("queue_signals.full_sweep.enqueued", float64(totalEnqueued), tags)
	e.mw.Gauge("queue_signals.full_sweep.errors", float64(totalErrors), tags)
	e.mw.Timing("queue_signals.full_sweep.duration_ms", duration, tags)

	e.l.Info("full sweep completed",
		zap.Int("enqueued", totalEnqueued),
		zap.Int("errors", totalErrors),
		zap.Duration("duration", duration))

	return &FullSweepResponse{
		Enqueued:   totalEnqueued,
		Errors:     totalErrors,
		DurationMS: duration.Milliseconds(),
	}, nil
}
