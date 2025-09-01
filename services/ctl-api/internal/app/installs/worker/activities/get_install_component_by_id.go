package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"gorm.io/gorm"
)

// @temporal-gen activity
func (a *Activities) GetInstallComponentByID(ctx context.Context, id string) (*app.InstallComponent, error) {
	installComponent, err := a.getInstallComponentByID(ctx, id)
	if err != nil {
		return nil, generics.TemporalGormError(err)
	}
	return installComponent, nil
}

func (a *Activities) getInstallComponentByID(ctx context.Context, installComponentID string) (*app.InstallComponent, error) {
	installComponent := app.InstallComponent{}
	res := a.db.WithContext(ctx).
		Preload("TerraformWorkspace").
		Preload("Component").
		Preload("Component.ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(10)
		}).
		First(&installComponent, "id = ?", installComponentID)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	return &installComponent, nil
}
