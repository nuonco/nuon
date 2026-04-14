package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field QueueID
func (a *Activities) getQueueSignals(ctx context.Context, queueID string) ([]*app.QueueSignal, error) {
	var queueSignals []*app.QueueSignal

	// Fetch signals that are queued, in-progress, or have no status set yet (empty JSONB).
	if res := a.db.WithContext(ctx).
		Where("queue_id = ?", queueID).
		Where("status->>'status' IN ? OR status->>'status' IS NULL", []string{string(app.StatusQueued), string(app.StatusInProgress)}).
		Order("created_at asc").
		Find(&queueSignals); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get queue signals")
	}

	return queueSignals, nil
}
