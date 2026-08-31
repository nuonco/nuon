package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateJobStatusRequest struct {
	JobID             string              `validate:"required"`
	Status            app.RunnerJobStatus `validate:"required"`
	StatusDescription string              `validate:"required"`
	ExpectedStatus    app.RunnerJobStatus
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateJobStatus(ctx context.Context, req *UpdateJobStatusRequest) error {
	job := app.RunnerJob{
		ID: req.JobID,
	}
	query := a.db.WithContext(ctx).Model(&job)
	if req.ExpectedStatus != "" {
		query = query.Where(app.RunnerJob{Status: req.ExpectedStatus})
	}
	res := query.Updates(app.RunnerJob{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update job status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		if req.ExpectedStatus != "" {
			return nil
		}
		return fmt.Errorf("no job found: %s %w", req.JobID, gorm.ErrRecordNotFound)
	}

	return nil
}
