package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// @temporal-gen-v2 activity
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field queueID
// @start-to-close-timeout 1m
func (a *Activities) queueExists(ctx context.Context, queueID string) (bool, error) {
	var count int64
	if res := a.db.WithContext(ctx).
		Model(&app.Queue{}).
		Where(app.Queue{ID: queueID}).
		Limit(1).
		Count(&count); res.Error != nil {
		return false, generics.TemporalGormError(res.Error, "unable to check queue existence")
	}

	return count > 0, nil
}
