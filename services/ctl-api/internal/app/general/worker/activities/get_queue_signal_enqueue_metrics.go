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

	// Live rows (deleted_at = 0) from planner stats. An exact count scans the
	// whole heap and exceeds the activity deadline on this table. ForceReplica:
	// background metrics cron, so replica lag is fine, and raw SQL has no
	// resolvable table for the allow-list to match.
	res := db.WithContext(ctx).
		Scopes(scopes.ForceReplica).
		Raw(`SELECT round(c.reltuples * m.freq)::bigint
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN pg_stats st
  ON st.schemaname = n.nspname
 AND st.tablename = c.relname
 AND st.attname = 'deleted_at'
CROSS JOIN LATERAL unnest(
  st.most_common_vals::text::text[],
  st.most_common_freqs
) AS m(val, freq)
WHERE n.nspname = 'public'
  AND c.relname = 'queue_signals'
  AND m.val = '0'`).
		Scan(&m.TotalQueued)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to estimate live queue signals: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("unable to estimate live queue signals: planner stats for deleted_at are missing")
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
