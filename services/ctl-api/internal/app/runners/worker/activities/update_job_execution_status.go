package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateJobExecutionStatusRequest struct {
	JobExecutionID string                       `validate:"required"`
	Status         app.RunnerJobExecutionStatus `validate:"required"`
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) UpdateJobExecutionStatus(ctx context.Context, req UpdateJobExecutionStatusRequest) error {
	runner := app.RunnerJobExecution{
		ID: req.JobExecutionID,
	}
	res := a.db.WithContext(ctx).Model(&runner).Updates(app.RunnerJobExecution{
		Status: req.Status,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update job execution status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no job execution found: %s %w", req.JobExecutionID, gorm.ErrRecordNotFound)
	}

	return nil
}
