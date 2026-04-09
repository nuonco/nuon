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
func (a *Activities) updateQueuePaused(ctx context.Context, queueID string, paused bool) error {
	if res := a.db.WithContext(ctx).Model(&app.Queue{}).Where("id = ?", queueID).Update("paused", paused); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to update queue paused state")
	}
	return nil
}
