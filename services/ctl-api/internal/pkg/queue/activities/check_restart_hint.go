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
func (a *Activities) checkRestartHint(ctx context.Context, queueID string) (bool, error) {
	var queue app.Queue
	if res := a.db.WithContext(ctx).Where(app.Queue{ID: queueID}).First(&queue); res.Error != nil {
		return false, generics.TemporalGormError(res.Error, "unable to get queue for restart hint check")
	}

	if queue.Metadata == nil {
		return false, nil
	}

	hint, ok := queue.Metadata["restart_hint"]
	if !ok || hint == nil {
		return false, nil
	}

	return true, nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
func (a *Activities) clearRestartHint(ctx context.Context, queueID string) error {
	var queue app.Queue
	if res := a.db.WithContext(ctx).Where(app.Queue{ID: queueID}).First(&queue); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to get queue for restart hint clear")
	}

	if queue.Metadata == nil {
		return nil
	}

	if _, ok := queue.Metadata["restart_hint"]; !ok {
		return nil
	}

	delete(queue.Metadata, "restart_hint")

	if res := a.db.WithContext(ctx).Model(&app.Queue{}).Where(app.Queue{ID: queueID}).Update("metadata", queue.Metadata); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to clear restart hint")
	}

	return nil
}
