package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type FlushOrphanedJobsRequest struct {
	RunnerID  string    `validate:"required"`
	Threshold time.Time `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) FlushOrphanedJobs(ctx context.Context, req FlushOrphanedJobsRequest) error {
	// dual-write V2 status
	compositeStatus := app.NewCompositeStatus(ctx, app.Status(app.RunnerJobStatusCancelled))

	res := a.db.WithContext(ctx).
		Model(&app.RunnerJob{}).
		Where(app.RunnerJob{
			RunnerID: req.RunnerID,
		}).
		Where("created_at < ?", req.Threshold).
		Where("status in (?)", []app.RunnerJobStatus{
			app.RunnerJobStatusQueued,
			app.RunnerJobStatusInProgress,
		}).
		Updates(map[string]any{
			"status":    app.RunnerJobStatusCancelled,
			"status_v2": compositeStatus,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to cancel orphaned jobs")
	}

	return nil
}
