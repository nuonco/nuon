package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type GetInstallComponentRequest struct {
	InstallID   string `validate:"required"`
	ComponentID string `validate:"required"`
}

func (a *Activities) GetInstallComponent(ctx context.Context, req GetInstallComponentRequest) (*app.InstallComponent, error) {
	installComponent := app.InstallComponent{}
	res := a.db.WithContext(ctx).
		Where(app.InstallComponent{
			InstallID:   req.InstallID,
			ComponentID: req.ComponentID,
		}).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at DESC")
		}).
		First(&installComponent)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install component: %w", res.Error)
	}

	return &installComponent, nil
}
