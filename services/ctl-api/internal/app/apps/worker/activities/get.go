package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRequest struct {
	AppID string `validate:"required"`
}

func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.App, error) {
	currentApp := app.App{}
	res := a.db.WithContext(ctx).
		Preload("Org").
		Preload("Installs").
		Preload("Installs.InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC").Limit(1)
		}).
		Preload("Components").
		Preload("CreatedBy").
		Preload("AppConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_configs_view_v1.created_at DESC").Limit(1)
		}).
		First(&currentApp, "id = ?", req.AppID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &currentApp, nil
}
