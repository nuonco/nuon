package activities

import (
	"context"
	"encoding/json"
	"fmt"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type GetInstallsVCSConfigOutput struct {
	VCSConfigID  string `json:"vcs_config_id"`
	InstallsDir  string `json:"installs_dir"`
	HasVCSConfig bool   `json:"has_vcs_config"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field appID
func (a *Activities) getInstallsVCSConfig(ctx context.Context, appID string) (*GetInstallsVCSConfigOutput, error) {
	var latestConfig app.AppConfig
	if err := a.db.WithContext(ctx).
		Where(app.AppConfig{AppID: appID}).
		Order("created_at DESC").
		First(&latestConfig).Error; err != nil {
		return nil, fmt.Errorf("unable to get latest app config: %w", err)
	}

	if latestConfig.IntermediateConfig == nil || !latestConfig.IntermediateConfig.IsSet() {
		return &GetInstallsVCSConfigOutput{}, nil
	}

	blobCtx := blobstore.WithBlobService(ctx, a.blobSvc)
	intermediateJSON, err := latestConfig.IntermediateConfig.Get(blobCtx)
	if err != nil || intermediateJSON == "" {
		return &GetInstallsVCSConfigOutput{}, nil
	}

	var parsed pkgconfig.AppConfig
	if err := json.Unmarshal([]byte(intermediateJSON), &parsed); err != nil {
		return nil, fmt.Errorf("unable to parse intermediate config: %w", err)
	}

	if parsed.InstallsConfig == nil {
		return &GetInstallsVCSConfigOutput{}, nil
	}

	ic := parsed.InstallsConfig

	if ic.ConnectedRepo != nil {
		vcsConfig, err := a.resolveConnectedVCSConfigID(ctx, appID, ic.ConnectedRepo.Repo, ic.ConnectedRepo.Branch)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve connected VCS config: %w", err)
		}
		return &GetInstallsVCSConfigOutput{
			VCSConfigID:  vcsConfig,
			InstallsDir:  ic.ConnectedRepo.Directory,
			HasVCSConfig: true,
		}, nil
	}

	if ic.PublicRepo != nil {
		vcsConfig, err := a.resolvePublicVCSConfigID(ctx, appID, ic.PublicRepo.Repo, ic.PublicRepo.Branch)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve public VCS config: %w", err)
		}
		return &GetInstallsVCSConfigOutput{
			VCSConfigID:  vcsConfig,
			InstallsDir:  ic.PublicRepo.Directory,
			HasVCSConfig: true,
		}, nil
	}

	return &GetInstallsVCSConfigOutput{}, nil
}

func (a *Activities) resolveConnectedVCSConfigID(ctx context.Context, appID, repo, branch string) (string, error) {
	var cfg app.ConnectedGithubVCSConfig
	err := a.db.WithContext(ctx).
		Where("repo = ?", repo).
		Order("created_at DESC").
		First(&cfg).Error
	if err != nil {
		return "", fmt.Errorf("connected VCS config not found for repo=%s: %w", repo, err)
	}
	return cfg.ID, nil
}

func (a *Activities) resolvePublicVCSConfigID(ctx context.Context, appID, repo, branch string) (string, error) {
	var cfg app.PublicGitVCSConfig
	err := a.db.WithContext(ctx).
		Where("repo = ?", repo).
		Order("created_at DESC").
		First(&cfg).Error
	if err != nil {
		return "", fmt.Errorf("public VCS config not found for repo=%s: %w", repo, err)
	}
	return cfg.ID, nil
}
