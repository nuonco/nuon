package sandbox

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

// Sync creates the app sandbox configuration via the shared builder in
// internal/pkg/config/build, which the CreateAppSandboxConfig handler also uses.
func Sync(ctx context.Context, db *gorm.DB, vcsHelper *vcshelpers.Helpers, cfg *config.AppConfig, appID, appConfigID string, state *sync.State) error {
	if cfg.Sandbox == nil {
		return sync.SyncErr{
			Resource:    "app-sandbox",
			Description: "sandbox config is required",
		}
	}

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

	in := build.SandboxInputFromConfig(cfg.Sandbox, appID, appConfigID)
	in.GithubVCSConfig = githubVCSConfig
	in.PublicGitVCSConfig = publicGitConfig

	appSandboxConfig, err := build.SandboxConfig(in)
	if err != nil {
		return sync.SyncErr{
			Resource:    "app-sandbox",
			Description: err.Error(),
		}
	}

	if res := db.WithContext(ctx).Create(appSandboxConfig); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app sandbox config",
			Err:         fmt.Errorf("unable to create app sandbox config: %w", res.Error),
		}
	}

	state.SandboxConfigID = appSandboxConfig.ID
	return nil
}
