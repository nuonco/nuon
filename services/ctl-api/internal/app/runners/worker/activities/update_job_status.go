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
}

// @temporal-gen-v2 activity
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
		return fmt.Errorf("no job found: %s %w", req.JobID, gorm.ErrRecordNotFound)
	}

	// dual-write V2 status
	compositeStatus := app.NewCompositeStatus(ctx, app.Status(req.Status))
	compositeStatus.StatusHumanDescription = req.StatusDescription
	a.db.WithContext(ctx).Model(&app.RunnerJob{ID: req.JobID}).Updates(map[string]any{
		"status_v2": compositeStatus,
	})

	return nil
}
