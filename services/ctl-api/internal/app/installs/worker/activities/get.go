package activities

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

// @temporal-gen-v2 activity
// @as-wrapper
// @by-field installID
func (a *Activities) get(ctx context.Context, installID string) (*app.Install, error) {
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

func (a *Activities) getInstall(ctx context.Context, installID string) (*app.Install, error) {
	return a.get(ctx, installID)
}

// SlimInstallResponse is a trimmed projection of app.Install for hot paths that
// only need core columns, avoiding confusion with a fully-preloaded install.
type SlimInstallResponse struct {
	ID          string
	OrgID       string
	AppID       string
	AppConfigID string
	SandboxMode sql.NullBool
	Metadata    pgtype.Hstore
}

// @temporal-gen-v2 activity
// @as-wrapper
// @by-field installID
func (a *Activities) getSlimInstall(ctx context.Context, installID string) (*SlimInstallResponse, error) {
	// full install object is quite costly from query and from logistics in temporal pov, this trim down version only
	// returns the metadat which is needed in application flow rather than entire app config
	install := app.Install{}
	res := a.db.WithContext(ctx).
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get install: %w")
	}

	return &SlimInstallResponse{
		ID:          install.ID,
		OrgID:       install.OrgID,
		AppID:       install.AppID,
		AppConfigID: install.AppConfigID,
		SandboxMode: install.SandboxMode,
		Metadata:    install.Metadata,
	}, nil
}
