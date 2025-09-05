package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetLatestDeployRequest struct {
	ComponentID string `json:"component_root_id" validate:"required"`
	InstallID   string `json:"install_id" validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetLatestDeploy(ctx context.Context, req GetLatestDeployRequest) (*app.InstallDeploy, error) {
	installCmpDeploy, err := a.getLatestDeploy(ctx, req.InstallID, req.ComponentID)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest deploy: %w", err)
	}

	return installCmpDeploy, nil
}

func (a *Activities) getLatestDeploy(ctx context.Context, installID, componentID string) (*app.InstallDeploy, error) {
	installCmp := app.InstallComponent{}
	res := a.db.WithContext(ctx).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Where(app.InstallDeploy{
                                Type: app.InstallDeployTypeApply,
                        }).Order("install_deploys.created_at DESC").Limit(1)
		}).
		Where(&app.InstallComponent{
			InstallID:   installID,
			ComponentID: componentID,
		}).
		First(&installCmp)
	if res.Error != nil && res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}
	if len(installCmp.InstallDeploys) != 1 {
		return nil, fmt.Errorf("no deploy exists for install: %w", gorm.ErrRecordNotFound)
	}

	if len(installCmp.InstallDeploys) == 0 {
		return nil, nil
	}

	return &installCmp.InstallDeploys[0], nil
}
