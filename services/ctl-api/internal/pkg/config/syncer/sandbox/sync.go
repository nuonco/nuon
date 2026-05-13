package sandbox

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/validation"
)

// Sync creates the app sandbox configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_sandbox_config.go
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string, state *sync.State) error {
	if cfg.Sandbox == nil {
		return sync.SyncErr{
			Resource:    "app-sandbox",
			Description: "sandbox config is required",
		}
	}

	if cfg.Sandbox.MaxAutoRetries != nil {
		if err := validation.ValidateMaxAutoRetries(*cfg.Sandbox.MaxAutoRetries); err != nil {
			return err
		}
	}

	// Get the app with preloaded org and VCS connections
	var parentApp app.App
	res := db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to get app",
			Err:         fmt.Errorf("unable to get app sandbox: %w", res.Error),
		}
	}

	// Build VCS configs using helpers
	vcsHelper := vcshelpers.New(vcshelpers.Params{})

	var githubVCSConfig *app.ConnectedGithubVCSConfig
	var publicGitConfig *app.PublicGitVCSConfig
	var err error

	if cfg.Sandbox.ConnectedRepo != nil {
		githubVCSConfig, err = vcsHelper.BuildConnectedGithubVCSConfig(ctx, &vcshelpers.ConnectedGithubVCSConfigRequest{
			Repo:      cfg.Sandbox.ConnectedRepo.Repo,
			Branch:    cfg.Sandbox.ConnectedRepo.Branch,
			Directory: cfg.Sandbox.ConnectedRepo.Directory,
		}, parentApp.Org)
		if err != nil {
			return sync.SyncInternalErr{
				Description: "unable to create connected github vcs config",
				Err:         fmt.Errorf("unable to create connected github vcs config: %w", err),
			}
		}
	}

	if cfg.Sandbox.PublicRepo != nil {
		publicGitConfig, err = vcsHelper.BuildPublicGitVCSConfig(ctx, &vcshelpers.PublicGitVCSConfigRequest{
			Repo:      cfg.Sandbox.PublicRepo.Repo,
			Branch:    cfg.Sandbox.PublicRepo.Branch,
			Directory: cfg.Sandbox.PublicRepo.Directory,
		})
		if err != nil {
			return sync.SyncInternalErr{
				Description: "unable to get public git config",
				Err:         fmt.Errorf("unable to get public git config: %w", err),
			}
		}
	}

	// Convert variables to pgtype.Hstore
	variables := make(map[string]*string)
	for k, v := range cfg.Sandbox.VarsMap {
		val := v
		variables[k] = &val
	}

	envVars := make(map[string]*string)
	for k, v := range cfg.Sandbox.EnvVarMap {
		val := v
		envVars[k] = &val
	}

	// Build variables files list
	variablesFiles := make([]string, 0)
	for _, vf := range cfg.Sandbox.VariablesFiles {
		variablesFiles = append(variablesFiles, vf.Contents)
	}

	// Build references list
	references := make([]string, 0)
	for _, ref := range cfg.Sandbox.References {
		references = append(references, ref.String())
	}

	appSandboxConfig := app.AppSandboxConfig{
		AppID:                    appID,
		AppConfigID:              appConfigID,
		PublicGitVCSConfig:       publicGitConfig,
		ConnectedGithubVCSConfig: githubVCSConfig,
		Variables:                pgtype.Hstore(variables),
		EnvVars:                  pgtype.Hstore(envVars),
		VariablesFiles:           pq.StringArray(variablesFiles),
		TerraformVersion:         cfg.Sandbox.TerraformVersion,
		References:               pq.StringArray(references),
		MaxAutoRetries:           cfg.Sandbox.MaxAutoRetries,
	}

	if cfg.Sandbox.DriftSchedule != nil {
		appSandboxConfig.DriftSchedule = *cfg.Sandbox.DriftSchedule
	}

	res = db.WithContext(ctx).Create(&appSandboxConfig)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app sandbox config",
			Err:         fmt.Errorf("unable to create app sandbox config: %w", res.Error),
		}
	}

	state.SandboxConfigID = appSandboxConfig.ID
	return nil
}
