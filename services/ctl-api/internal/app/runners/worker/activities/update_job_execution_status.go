package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateJobExecutionStatusRequest struct {
	JobExecutionID string                       `validate:"required"`
	Status         app.RunnerJobExecutionStatus `validate:"required"`
}

// @temporal-gen-v2 activity
// @max-retries 2
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
		return generics.TemporalGormError(
			gorm.ErrRecordNotFound,
			fmt.Sprintf("no job execution found: %s", req.JobExecutionID),
		)
	}
	helpers.AuditJobExecutionResult(ctx, a.db, a.mw, req.JobExecutionID, req.Status, "runner_workflow")

	return nil
}
