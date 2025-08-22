package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateDeployStatusRequest struct {
	DeployID          string                  `validate:"required"`
	Status            app.InstallDeployStatus `validate:"required"`
	StatusDescription string                  `validate:"required"`
	SkipStatusSync    bool
}

// @temporal-gen activity
func (a *Activities) UpdateDeployStatus(ctx context.Context, req UpdateDeployStatusRequest) error {
	installDeploy := app.InstallDeploy{
		ID: req.DeployID,
	}
	res := a.db.WithContext(ctx).Model(&installDeploy).Updates(app.InstallDeploy{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install deploy: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no install found: %s %w", req.DeployID, gorm.ErrRecordNotFound)
	}

	extantInstallDeploy := app.InstallDeploy{}
	res = a.db.WithContext(ctx).
		Preload("InstallComponent").
		Where("id = ?", req.DeployID).
		First(&extantInstallDeploy)
	if res.Error != nil {
		return fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	if !req.SkipStatusSync {
		installComponent := app.InstallComponent{
			ID: extantInstallDeploy.InstallComponent.ID,
		}
		res = a.db.WithContext(ctx).
			Model(&installComponent).
			Updates(app.InstallComponent{
				Status:            app.DeployStatusToComponentStatus(req.Status),
				StatusDescription: req.StatusDescription,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update install component: %w", res.Error)
		}
	}

	return nil
}
