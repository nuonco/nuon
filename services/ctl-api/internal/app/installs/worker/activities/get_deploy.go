package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type GetDeployRequest struct {
	DeployID string `validate:"required"`
}

func (a *Activities) GetDeploy(ctx context.Context, req GetDeployRequest) (*app.InstallDeploy, error) {
	installDeploy := app.InstallDeploy{}
	res := a.db.WithContext(ctx).
		Preload("ComponentBuild").
		Preload("InstallComponent").
		Preload("InstallComponent.Install").
		Preload("InstallComponent.Install.InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs.created_at DESC")
		}).
		Preload("InstallComponent.Install.InstallInputs.AppInputConfig").
		Preload("InstallComponent.Install.AppSandboxConfig").
		Preload("InstallComponent.Install.AppRunnerConfig").
		First(&installDeploy, "id = ?", req.DeployID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	return &installDeploy, nil
}
