package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type UpdateReleaseStepStatusRequest struct {
	ReleaseStepID     string `validate:"required"`
	Status            string `validate:"required"`
	StatusDescription string `validate:"required"`
}

func (a *Activities) UpdateReleaseStepStatus(ctx context.Context, req UpdateReleaseStepStatusRequest) error {
	release := app.ComponentReleaseStep{
		ID: req.ReleaseStepID,
	}
	res := a.db.WithContext(ctx).Model(&release).Updates(app.ComponentReleaseStep{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update release step: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no release step found: %s %w", req.ReleaseStepID, gorm.ErrRecordNotFound)
	}

	return nil
}
