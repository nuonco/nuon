package activities

import (
	"context"
	"fmt"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
// @as-wrapper
// @by-field appID
func (a *Activities) getLatestAppSandboxConfig(ctx context.Context, appID string) (*app.AppSandboxConfig, error) {
	var cfg app.AppSandboxConfig
	res := a.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("PublicGitVCSConfig").
		Where("app_id = ?", appID).
		Order("created_at DESC").
		First(&cfg)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app sandbox config for app %s: %w", appID, res.Error)
	}
	return &cfg, nil
}

type GetSandboxBuildGitSourceRequest struct {
	SandboxConfigID string `json:"sandbox_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetSandboxBuildGitSource(ctx context.Context, req GetSandboxBuildGitSourceRequest) (*plantypes.GitSource, error) {
	var cfg app.AppSandboxConfig
	res := a.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("PublicGitVCSConfig").
		First(&cfg, "id = ?", req.SandboxConfigID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox config: %w", res.Error)
	}

	switch cfg.VCSConnectionType {
	case app.VCSConnectionTypeConnectedRepo:
		return a.vcsHelpers.GetGitSource(ctx, cfg.ConnectedGithubVCSConfig)
	case app.VCSConnectionTypePublicRepo:
		return a.vcsHelpers.GetPubliGitSource(ctx, cfg.PublicGitVCSConfig)
	default:
		return nil, fmt.Errorf("sandbox config %s has no VCS config", req.SandboxConfigID)
	}
}
