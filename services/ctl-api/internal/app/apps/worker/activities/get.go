package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

type GetRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen activity
// @by-id AppID
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
			return db.Order(views.TableOrViewName(a.db, &app.AppConfig{}, ".created_at DESC")).Limit(1)
		}).
		First(&currentApp, "id = ?", req.AppID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &currentApp, nil
}
