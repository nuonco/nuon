package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) GetInstallComponent(ctx context.Context, installID, componentID string) (*app.InstallComponent, error) {
	installCmp := app.InstallComponent{}
	res := h.db.WithContext(ctx).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.
				Order("install_deploys.created_at DESC").Limit(1)
		}).
		Preload("TerraformWorkspace").
		Where(&app.InstallComponent{
			InstallID:   installID,
			ComponentID: componentID,
		}).
		First(&installCmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install component: %w", res.Error)
	}

	return &installCmp, nil
}
