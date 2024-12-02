package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) GetInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := h.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("AppSandboxConfig").
		Preload("AppRunnerConfig").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs_view_v1.created_at DESC")
		}).
		Preload("App").
		Preload("App.Org").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
