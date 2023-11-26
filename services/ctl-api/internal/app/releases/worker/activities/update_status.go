package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type UpdateStatusRequest struct {
	ReleaseID         string `validate:"required"`
	Status            string `validate:"required"`
	StatusDescription string `validate:"required"`
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
