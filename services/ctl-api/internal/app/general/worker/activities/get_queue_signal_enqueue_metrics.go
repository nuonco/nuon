package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type UnenqueuedByType struct {
	Type      string `gorm:"column:type"`
	Namespace string `gorm:"column:namespace"`
	Count     int64  `gorm:"column:count"`
}

type QueueSignalEnqueueMetrics struct {
	TotalQueued          int64
	MissingEnqueueFinish int64
	UnenqueuedByType     []UnenqueuedByType
}

type GetQueueSignalEnqueueMetricsRequest struct{}

// @temporal-gen-v2 activity
func (a *Activities) GetQueueSignalEnqueueMetrics(ctx context.Context, req GetQueueSignalEnqueueMetricsRequest) (*QueueSignalEnqueueMetrics, error) {
	return a.getQueueSignalEnqueueMetrics(ctx, a.db)
}

func (a *Activities) getQueueSignalEnqueueMetrics(ctx context.Context, db *gorm.DB) (*QueueSignalEnqueueMetrics, error) {
	var m QueueSignalEnqueueMetrics

	// Count total queue signals in non-terminal queued state. Force this read
	// onto the replica pool: it's a background metrics cron, so replica lag is
	// irrelevant, and ForceReplica bypasses the table ACL (raw SQL has no
	// resolvable table for the allow-list to match).
	if res := db.WithContext(ctx).
		Scopes(scopes.ForceReplica).
		Raw(`SELECT count(*) FROM queue_signals WHERE deleted_at = 0`).
		Scan(&m.TotalQueued); res.Error != nil {
		return nil, fmt.Errorf("unable to count total queue signals: %w", res.Error)
	}

	// Count queue signals that have not been enqueued yet.
	if res := db.WithContext(ctx).
		Raw(`SELECT count(*) FROM queue_signals WHERE deleted_at = 0 AND enqueued = false`).
		Scan(&m.MissingEnqueueFinish); res.Error != nil {
		return nil, fmt.Errorf("unable to count signals not yet enqueued: %w", res.Error)
	}

	// Break down unenqueued signals by type for per-signal-type gauges.
	if res := db.WithContext(ctx).
		Raw(`SELECT qs.type, q.workflow->>'namespace' as namespace, count(*) as count FROM queue_signals qs JOIN queues q ON q.id = qs.queue_id WHERE qs.deleted_at = 0 AND qs.enqueued = false GROUP BY qs.type, q.workflow->>'namespace'`).
		Scan(&m.UnenqueuedByType); res.Error != nil {
		return nil, fmt.Errorf("unable to count unenqueued signals by type: %w", res.Error)
	}

	return &m, nil
}
