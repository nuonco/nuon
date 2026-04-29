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

	if queue.StatusV2.Metadata == nil {
		return false, nil
	}

	val, ok := queue.StatusV2.Metadata["restart_hint"]
	if !ok || val == nil {
		return false, nil
	}

	return true, nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
func (a *Activities) clearRestartHint(ctx context.Context, queueID string) error {
	if err := generics.MergeJSONBMetadata(a.db.WithContext(ctx), &app.Queue{}, queueID, "status_v2", map[string]any{
		"restart_hint": nil,
	}); err != nil {
		return generics.TemporalGormError(err, "unable to clear restart hint")
	}

	return nil
}
