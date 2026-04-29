package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type QueueSignalEnqueueMetrics struct {
	TotalQueued          int64
	MissingEnqueueFinish int64
}

type GetQueueSignalEnqueueMetricsRequest struct{}

// @temporal-gen-v2 activity
func (a *Activities) GetQueueSignalEnqueueMetrics(ctx context.Context, req GetQueueSignalEnqueueMetricsRequest) (*QueueSignalEnqueueMetrics, error) {
	return a.getQueueSignalEnqueueMetrics(ctx, a.db)
}

func (a *Activities) getQueueSignalEnqueueMetrics(ctx context.Context, db *gorm.DB) (*QueueSignalEnqueueMetrics, error) {
	var m QueueSignalEnqueueMetrics

	// Count total queue signals in non-terminal queued state
	if res := db.WithContext(ctx).
		Raw(`SELECT count(*) FROM queue_signals WHERE deleted_at = 0`).
		Scan(&m.TotalQueued); res.Error != nil {
		return nil, fmt.Errorf("unable to count total queue signals: %w", res.Error)
	}

	// Count queue signals missing enqueue_finished_at in their status metadata.
	// These are signals where the background enqueue goroutine either hasn't
	// completed yet or failed silently.
	if res := db.WithContext(ctx).
		Raw(`SELECT count(*) FROM queue_signals
WHERE deleted_at = 0
AND NOT (status::jsonb -> 'metadata' ? 'enqueue_finished_at')`).
		Scan(&m.MissingEnqueueFinish); res.Error != nil {
		return nil, fmt.Errorf("unable to count signals missing enqueue_finished_at: %w", res.Error)
	}

	return &m, nil
}
