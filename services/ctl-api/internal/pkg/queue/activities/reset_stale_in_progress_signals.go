package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type ResetStaleInProgressSignalsRequest struct {
	QueueID string `json:"queue_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ResetStaleInProgressSignals(ctx context.Context, req *ResetStaleInProgressSignalsRequest) error {
	res := a.db.WithContext(ctx).
		Model(&app.QueueSignal{}).
		Scopes(generics.WhereJSONBStatus("status", string(app.StatusInProgress))).
		Where(app.QueueSignal{
			QueueID:  req.QueueID,
			Enqueued: true,
		}).Update("status", map[string]any{
		"status": string(app.StatusQueued),
	})
	if res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to reset stale in-progress signals")
	}

	return nil
}
