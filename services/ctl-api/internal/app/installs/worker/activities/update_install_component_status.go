package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateInstallComponentStatusRequest struct {
	InstallComponentID string                     `validate:"required"`
	Status             app.InstallComponentStatus `validate:"required"`
	StatusDescription  string                     `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateInstallComponentStatus(ctx context.Context, req UpdateInstallComponentStatusRequest) error {
	installComponent := app.InstallComponent{
		ID: req.InstallComponentID,
	}
	res := a.db.WithContext(ctx).
		Model(&installComponent).
		Updates(app.InstallComponent{
			Status:            req.Status,
			StatusDescription: req.StatusDescription,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update install component: %w", res.Error)
	}

	if res.RowsAffected < 1 {
		return fmt.Errorf("no install found: %s %w", req.InstallComponentID, gorm.ErrRecordNotFound)
	}

	return nil
}
