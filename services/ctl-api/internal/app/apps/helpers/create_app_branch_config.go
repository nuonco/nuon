package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateAppBranchConfigOptions struct {
	InstallsConnectedGithubVCSConfig *app.ConnectedGithubVCSConfig
	InstallsPublicGitVCSConfig       *app.PublicGitVCSConfig
}

func (h *Helpers) CreateAppBranchConfig(
	ctx context.Context,
	appBranchID string,
	connectedGithubVCSConfig *app.ConnectedGithubVCSConfig,
	publicGitVCSConfig *app.PublicGitVCSConfig,
	installGroups []app.AppBranchInstallGroup,
	opts *CreateAppBranchConfigOptions,
) (*app.AppBranchConfig, error) {
	config := app.AppBranchConfig{
		AppBranchID:              appBranchID,
		InstallGroups:            installGroups,
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,
		PublicGitVCSConfig:       publicGitVCSConfig,
	}

	if opts != nil {
		config.InstallsConnectedGithubVCSConfig = opts.InstallsConnectedGithubVCSConfig
		config.InstallsPublicGitVCSConfig = opts.InstallsPublicGitVCSConfig
	}

	// Create config - GORM will automatically create VCS configs due to polymorphic relationship
	if err := h.db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, fmt.Errorf("unable to create app branch config: %w", err)
	}

	return &config, nil
}
