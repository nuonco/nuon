package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetInstallComponentRequest struct {
	InstallID   string `validate:"required"`
	ComponentID string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetInstallComponent(ctx context.Context, req GetInstallComponentRequest) (*app.InstallComponent, error) {
	installComponent, err := a.getInstallComponent(ctx, req.InstallID, req.ComponentID)
	if err != nil {
		return nil, err
	}
	return installComponent, nil
}

func (a *Activities) getInstallComponent(ctx context.Context, installID, componentID string) (*app.InstallComponent, error) {
	installComponent := app.InstallComponent{}
	res := a.db.WithContext(ctx).
		Preload("Component").
		Preload("Component.Dependencies").
		Preload("TerraformWorkspace").
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at DESC").Limit(1)
		}).
		Where(app.InstallComponent{
			InstallID:   installID,
			ComponentID: componentID,
		}).
		First(&installComponent)
	if res.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install component: %w", res.Error)
	}

	return &installComponent, nil
}
