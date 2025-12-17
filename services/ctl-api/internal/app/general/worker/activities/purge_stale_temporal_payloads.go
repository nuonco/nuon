package activities

import (
	"context"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type PurgeStaleTemporalPayloadsRequest struct {
	DurationAgo time.Duration
}

// @temporal-gen activity
func (a *Activities) PurgeStaleTemporalPayloads(ctx context.Context, req PurgeStaleTemporalPayloadsRequest) error {
	cutoff := time.Now().Add(-req.DurationAgo)

	res := a.db.WithContext(ctx).
		Unscoped().
		Where("created_at < ?", cutoff).
		Delete(&app.TemporalPayload{})

	if res.Error != nil {
		return res.Error
	}

	a.mw.Count("event_loop.general.purge_stale_data", int64(res.RowsAffected), []string{})
	return nil
}
