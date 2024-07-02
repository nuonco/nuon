package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateStatusRequest struct {
	ReleaseID         string            `validate:"required"`
	Status            app.ReleaseStatus `validate:"required"`
	StatusDescription string            `validate:"required"`
}

func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	release := app.ComponentRelease{
		ID: req.ReleaseID,
	}
	res := a.db.WithContext(ctx).Model(&release).Updates(app.ComponentRelease{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update release: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no release found: %s %w", req.ReleaseID, gorm.ErrRecordNotFound)
	}
	return nil
}
