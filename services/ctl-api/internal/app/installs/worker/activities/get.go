package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) Get(ctx context.Context, req GetRequest) (*app.Install, error) {
	return a.getInstall(ctx, req.InstallID)
}

func (a *Activities) getInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := a.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Org").
		Preload("Org.RunnerGroup").
		Preload("Org.RunnerGroup.Runners").
		Preload("App").
		Preload("App.Org").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.AWSDelegationConfig").
		Preload("InstallSandbox").
		Preload("InstallSandbox.TerraformWorkspace").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs_view_v1.created_at DESC")
		}).
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC").Limit(1)
		}).

		// load app secrets for deploys
		Preload("App.AppSecrets").
		Preload("AppRunnerConfig").

		// load connected github
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig.VCSConnection").

		// load public git
		Preload("AppSandboxConfig.PublicGitVCSConfig").

		// load runners
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("RunnerGroup.Runners.RunnerGroup").
		First(&install, "id = ?", installID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
