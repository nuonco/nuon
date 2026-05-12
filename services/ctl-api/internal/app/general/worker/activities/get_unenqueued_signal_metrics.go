package activities

import (
	"context"
	"fmt"
)

type UnenqueuedSignalMetrics struct {
	// Total signals that are not enqueued and not in a terminal state.
	Total int64
	// Stale signals not enqueued and older than 5 minutes.
	Stale int64
}

type GetUnenqueuedSignalMetricsRequest struct{}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetUnenqueuedSignalMetrics(ctx context.Context, req GetUnenqueuedSignalMetricsRequest) (*UnenqueuedSignalMetrics, error) {
	m := &UnenqueuedSignalMetrics{}

	if res := a.db.WithContext(ctx).Raw(`
		SELECT count(*)
		FROM queue_signals
		WHERE deleted_at = 0
		AND enqueued = false
		AND status::jsonb ->> 'status' NOT IN ('success', 'error', 'cancelled')
	`).Scan(&m.Total); res.Error != nil {
		return nil, fmt.Errorf("unable to count unenqueued signals: %w", res.Error)
	}

	if res := a.db.WithContext(ctx).Raw(`
		SELECT count(*)
		FROM queue_signals
		WHERE deleted_at = 0
		AND enqueued = false
		AND status::jsonb ->> 'status' NOT IN ('success', 'error', 'cancelled')
		AND created_at < NOW() - INTERVAL '5 minutes'
	`).Scan(&m.Stale); res.Error != nil {
		return nil, fmt.Errorf("unable to count stale unenqueued signals: %w", res.Error)
	}

	return m, nil
}
