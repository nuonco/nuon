package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (h *Helpers) GetInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := h.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("AppSandboxConfig").
		Preload("AppRunnerConfig").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs.created_at DESC")
		}).
		Preload("App").
		Preload("App.Org").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil

}
