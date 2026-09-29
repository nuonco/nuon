package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) GetInstallComponents(ctx context.Context, installID string) ([]app.InstallComponent, error) {
	install := &app.Install{}
	res := h.db.WithContext(ctx).
		Preload("InstallComponents", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_components.created_at DESC")
		}).
		Preload("InstallComponents.InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.
				Where("status IN ?", app.AppliedDeployStatuses())
		}).
		Preload("InstallComponents.Component").
		Preload("InstallComponents.TerraformWorkspace").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install components: %w", res.Error)
	}

	return install.InstallComponents, nil
}
