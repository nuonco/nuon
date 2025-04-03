package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetDeployRequest struct {
	DeployID string `validate:"required"`
}

// @temporal-gen activity
// @by-id DeployID
func (a *Activities) GetDeploy(ctx context.Context, req GetDeployRequest) (*app.InstallDeploy, error) {
	return a.getDeploy(ctx, req.DeployID)
}

func (a *Activities) getDeploy(ctx context.Context, deployID string) (*app.InstallDeploy, error) {
	installDeploy := app.InstallDeploy{}
	res := a.db.WithContext(ctx).
		Preload("ComponentBuild").

		// load install
		Preload("InstallComponent").
		Preload("InstallComponent.Component").
		Preload("InstallComponent.Install").
		Preload("InstallComponent.Install.InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs_view_v1.created_at DESC")
		}).
		Preload("InstallComponent.Install.InstallInputs.AppInputConfig").
		Preload("InstallComponent.Install.InstallInputs").
		Preload("InstallComponent.Install.AppSandboxConfig").
		Preload("InstallComponent.Install.AppRunnerConfig").
		First(&installDeploy, "id = ?", deployID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install deploy: %w", res.Error)
	}

	return &installDeploy, nil
}
