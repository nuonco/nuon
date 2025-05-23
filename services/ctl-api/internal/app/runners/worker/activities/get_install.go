package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

type GetInstallRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetInstall(ctx context.Context, req GetInstallRequest) (*app.Install, error) {
	return a.getInstall(ctx, req.InstallID)
}

func (a *Activities) getInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := a.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("App").
		Preload("App.Org").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("RunnerGroup.Runners").
		Preload("RunnerGroup.Settings").
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("AppSandboxConfig.AWSDelegationConfig").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.InstallInputs{}, ".created_at DESC"))
		}).
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC")
		}).
		First(&install, "id = ?", installID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
