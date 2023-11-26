package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type UpdateStatusRequest struct {
	InstallID         string `validate:"required"`
	Status            string `validate:"required"`
	StatusDescription string `validate:"required"`
}

func (a *Activities) UpdateStatus(ctx context.Context, req UpdateStatusRequest) error {
	install := app.Install{
		ID: req.InstallID,
	}
	res := a.db.WithContext(ctx).Model(&install).Updates(app.Install{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no install found: %s %w", req.InstallID, gorm.ErrRecordNotFound)
	}
	return nil
}
