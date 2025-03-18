package helpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// GetInstall reads an install from the DB.
func (h *Helpers) GetInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := h.db.WithContext(ctx).
		Preload("AppRunnerConfig").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs_view_v1.created_at DESC").Limit(1)
		}).
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("App").
		Preload("App.AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC").Limit(1)
		}).
		Preload("App.AppSecrets", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_secrets.created_at DESC")
		}).
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC").Limit(5)
		}).
		Preload("InstallSandboxRuns.AppSandboxConfig").
		Preload("App.Org").
		Preload("AppSandboxConfig").
		Where("name = ?", installID).
		Or("id = ?", installID).
		First(&install)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install")
	}

	return &install, nil
}
