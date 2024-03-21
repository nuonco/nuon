package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type GetRequest struct {
	InstallID string `validate:"required"`
}

func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.Install, error) {
	return a.getInstall(ctx, req.InstallID)
}

func (a *Activities) getInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := a.db.WithContext(ctx).
		Preload("App").
		Preload("App.Org").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("AppSandboxConfig").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs.created_at DESC")
		}).
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC")
		}).

		// load sandbox
		Preload("AppSandboxConfig.SandboxRelease").
		Preload("AppSandboxConfig.SandboxRelease.Sandbox").
		Preload("AppRunnerConfig").

		// load connected github
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig.VCSConnection").

		// load public git
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		First(&install, "id = ?", installID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
