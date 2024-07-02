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
}

func (a *Activities) UpdateDeployStatus(ctx context.Context, req UpdateDeployStatusRequest) error {
	install := app.InstallDeploy{
		ID: req.DeployID,
	}
	res := a.db.WithContext(ctx).Model(&install).Updates(app.InstallDeploy{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update install deploy: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no install found: %s %w", req.DeployID, gorm.ErrRecordNotFound)
	}
	return nil
}
