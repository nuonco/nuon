package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateDeployStatusV2Request struct {
	DeployID          string     `validate:"required"`
	Status            app.Status `validate:"required"`
	StatusDescription string     `validate:"required"`
	SkipStatusSync    bool
}

// @temporal-gen activity
func (a *Activities) UpdateDeployStatusV2(ctx context.Context, req UpdateDeployStatusV2Request) error {
	installDeploy := app.InstallDeploy{
		ID: req.DeployID,
	}

	compositeStatus := app.NewCompositeStatus(ctx, req.Status)
	compositeStatus.StatusHumanDescription = req.StatusDescription

	res := a.db.WithContext(ctx).Model(&installDeploy).Updates(app.InstallDeploy{
		StatusV2: compositeStatus,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install deploy: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no install found: %s %w", req.DeployID, gorm.ErrRecordNotFound)
	}

	if req.SkipStatusSync {
		return nil
	}

	extantInstallDeploy := app.InstallDeploy{}
	res = a.db.WithContext(ctx).
		Preload("InstallComponent").
		Where("id = ?", req.DeployID).
		First(&extantInstallDeploy)
	if res.Error != nil {
		return fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	compositeStatus = app.NewCompositeStatus(ctx, req.Status)
	compositeStatus.StatusHumanDescription = req.StatusDescription
	installComponent := app.InstallComponent{
		ID: extantInstallDeploy.InstallComponent.ID,
	}
	res = a.db.WithContext(ctx).
		Model(&installComponent).
		Updates(app.InstallComponent{
			StatusV2: compositeStatus,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update install component: %w", res.Error)
	}

	return nil
}
