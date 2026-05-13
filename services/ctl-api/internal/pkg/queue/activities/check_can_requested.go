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
func (a *Activities) checkCANRequested(ctx context.Context, queueID string) (bool, error) {
	var queue app.Queue
	if res := a.db.WithContext(ctx).Where(app.Queue{ID: queueID}).First(&queue); res.Error != nil {
		return false, generics.TemporalGormError(res.Error, "unable to get queue for CAN check")
	}

	if queue.StatusV2.Metadata == nil {
		return false, nil
	}

	val, ok := queue.StatusV2.Metadata["restart_hint"]
	if !ok || val == nil {
		return false, nil
	}

	return true, nil
}

// clearCANRequested removes the restart_hint key from the queue's status_v2 metadata.
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
func (a *Activities) clearCANRequested(ctx context.Context, queueID string) error {
	res := a.db.WithContext(ctx).Exec(`
		UPDATE queues
		SET status_v2 = jsonb_set(
			COALESCE(status_v2::jsonb, '{}'::jsonb),
			'{metadata}',
			COALESCE(status_v2::jsonb -> 'metadata', '{}'::jsonb) - 'restart_hint'
		)
		WHERE id = ? AND deleted_at = 0
	`, queueID)
	if res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to clear restart_hint on queue")
	}
	return nil
}
