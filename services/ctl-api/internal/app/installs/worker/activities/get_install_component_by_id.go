package activities

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
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
		First(&installComponent, "id = ?", installComponentID)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	return &installComponent, nil
}
