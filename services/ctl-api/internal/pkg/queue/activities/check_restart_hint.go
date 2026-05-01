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
	// Look up the queue to find the owning runner.
	var queue app.Queue
	if res := a.db.WithContext(ctx).Where(app.Queue{ID: queueID}).First(&queue); res.Error != nil {
		return false, generics.TemporalGormError(res.Error, "unable to get queue for restart hint check")
	}

	if queue.OwnerType != "runners" || queue.OwnerID == "" {
		return false, nil
	}

	// Check the runner's dedicated column instead of queue JSONB metadata.
	var runner app.Runner
	if res := a.db.WithContext(ctx).Select("restart_requested").Where(app.Runner{ID: queue.OwnerID}).First(&runner); res.Error != nil {
		return false, generics.TemporalGormError(res.Error, "unable to get runner for restart hint check")
	}

	return runner.RestartRequested, nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
func (a *Activities) clearRestartHint(ctx context.Context, queueID string) error {
	// Look up the queue to find the owning runner.
	var queue app.Queue
	if res := a.db.WithContext(ctx).Where(app.Queue{ID: queueID}).First(&queue); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to get queue for clear restart hint")
	}

	if queue.OwnerType != "runners" || queue.OwnerID == "" {
		return nil
	}

	// Clear the runner's restart_requested column.
	if res := a.db.WithContext(ctx).Model(&app.Runner{}).Where(app.Runner{ID: queue.OwnerID}).Update("restart_requested", false); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to clear restart_requested on runner")
	}

	return nil
}
