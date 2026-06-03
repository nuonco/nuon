package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processhealthcheck"
)

const oldQueueSignalRetention = 7 * 24 * time.Hour

type DeleteOldQueueSignalsRequest struct {
	BatchSize int `json:"batch_size"`
}

type DeleteOldQueueSignalsResponse struct {
	Deleted int64 `json:"deleted"`
}

// DeleteOldQueueSignals hard-deletes a single batch of process_healthcheck and
// healthcheck queue signals older than the retention window, returning the rows
// removed so the caller can loop until a batch comes back smaller than BatchSize.
// It no-ops (returns 0) unless QueueSignalCleanupEnabled is set.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) DeleteOldQueueSignals(ctx context.Context, req DeleteOldQueueSignalsRequest) (*DeleteOldQueueSignalsResponse, error) {
	if !a.cfg.QueueSignalCleanupEnabled {
		return &DeleteOldQueueSignalsResponse{}, nil
	}
	if req.BatchSize <= 0 {
		req.BatchSize = 50000
	}

	cutoff := time.Now().Add(-oldQueueSignalRetention)

	sub := a.db.Model(&app.QueueSignal{}).
		Unscoped().
		Select("id").
		Where("type IN ?", []string{string(processhealthcheck.SignalType), "healthcheck"}).
		Where("created_at < ?", cutoff).
		Limit(req.BatchSize)

	res := a.db.WithContext(ctx).
		Unscoped().
		Where("id IN (?)", sub).
		Delete(&app.QueueSignal{})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to delete old queue signals: %w", res.Error)
	}

	a.mw.Count("general.queue_signal_cleanup.deleted", res.RowsAffected, []string{})
	return &DeleteOldQueueSignalsResponse{Deleted: res.RowsAffected}, nil
}
