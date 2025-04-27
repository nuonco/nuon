package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type FlushOrphanedJobsRequest struct {
	RunnerID  string    `validate:"required"`
	Threshold time.Time `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) FlushOrphanedJobs(ctx context.Context, req FlushOrphanedJobsRequest) error {
	res := a.db.WithContext(ctx).
		Where(app.RunnerJob{
			RunnerID: req.RunnerID,
		}).
		Where("created_at < ?", req.Threshold).
		Where("status in (?)", []app.RunnerJobStatus{
			app.RunnerJobStatusQueued,
			app.RunnerJobStatusInProgress,
		}).
		Updates(app.RunnerJob{
			Status: app.RunnerJobStatusCancelled,
		})
	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to cancel orphaned jobs")
	}

	return nil
}
