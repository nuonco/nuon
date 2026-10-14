package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type UpdateDeployStatusRequest struct {
	DeployID          string                  `validate:"required"`
	Status            app.InstallDeployStatus `validate:"required"`
	StatusDescription string                  `validate:"required"`
	SkipStatusSync    bool
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateDeployStatus(ctx context.Context, req UpdateDeployStatusRequest) error {
	installDeploy := app.InstallDeploy{
		ID: req.DeployID,
	}
	if req.Status == app.InstallDeployStatusCancelled {
		_, err := statusactivities.CancelIfUnfinished(ctx, a.db, &installDeploy, req.StatusDescription, req.SkipStatusSync)
		return err
	}
	query := a.db.WithContext(ctx).Model(&installDeploy).Where("status <> ?", app.InstallDeployStatusCancelled)
	res := query.Updates(app.InstallDeploy{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install deploy: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		var existing app.InstallDeploy
		if err := a.db.WithContext(ctx).Where(app.InstallDeploy{ID: req.DeployID}).First(&existing).Error; err == nil && existing.Status == app.InstallDeployStatusCancelled {
			return nil
		}
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
