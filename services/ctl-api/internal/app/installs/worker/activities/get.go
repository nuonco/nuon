package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type GetRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
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
		Preload("AppConfig").
		Preload("App.Org").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("GCPAccount").
		Preload("AppSandboxConfig").
		Preload("InstallSandbox").
		Preload("InstallSandbox.TerraformWorkspace").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.InstallInputs{}, ".created_at DESC"))
		}).
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC").Limit(1)
		}).

		// load app secrets for deploys
		Preload("App.AppSecrets").
		Preload("AppRunnerConfig").
		Preload("InstallConfig").

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
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get install: %w")
	}

	return &install, nil
}
