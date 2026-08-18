package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateJobStatusRequest struct {
	JobID             string              `validate:"required"`
	Status            app.RunnerJobStatus `validate:"required"`
	StatusDescription string              `validate:"required"`
}

// @temporal-gen-v2 activity
// @max-retries 2
func (a *Activities) UpdateJobStatus(ctx context.Context, req UpdateJobStatusRequest) error {
	runner := app.RunnerJob{
		ID: req.JobID,
	}
	res := a.db.WithContext(ctx).Model(&runner).Updates(app.RunnerJob{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update job status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(
			gorm.ErrRecordNotFound,
			fmt.Sprintf("no job found: %s", req.JobID),
		)
	}

	return nil
}
