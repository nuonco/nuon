package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type NotEnqueuedByType struct {
	Type      string `gorm:"column:type"`
	Namespace string `gorm:"column:namespace"`
	Count     int64  `gorm:"column:count"`
}

type QueueSignalNotEnqueuedMetrics struct {
	NotEnqueuedByType []NotEnqueuedByType
}

type GetQueueSignalNotEnqueuedMetricsRequest struct{}

// @temporal-gen-v2 activity
func (a *Activities) GetQueueSignalNotEnqueuedMetrics(ctx context.Context, req GetQueueSignalNotEnqueuedMetricsRequest) (*QueueSignalNotEnqueuedMetrics, error) {
	return a.getQueueSignalNotEnqueuedMetrics(ctx, a.db)
}

func (a *Activities) getQueueSignalNotEnqueuedMetrics(ctx context.Context, db *gorm.DB) (*QueueSignalNotEnqueuedMetrics, error) {
	var m QueueSignalNotEnqueuedMetrics

	if res := db.WithContext(ctx).
		Raw(`SELECT type, COALESCE(workflow->>'namespace', '') AS namespace, count(*) as count FROM queue_signals WHERE enqueued = false GROUP BY type, namespace`).
		Scan(&m.NotEnqueuedByType); res.Error != nil {
		return nil, fmt.Errorf("unable to count not enqueued signals by type: %w", res.Error)
	}

	return &m, nil
}
