package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field QueueID
func (a *Activities) updateQueuePaused(ctx context.Context, queueID string, paused bool) error {
	if res := a.db.WithContext(ctx).Model(&app.Queue{}).Where("id = ?", queueID).Update("paused", paused); res.Error != nil {
		return errors.Wrap(res.Error, "unable to update queue paused state")
	}
	return nil
}
