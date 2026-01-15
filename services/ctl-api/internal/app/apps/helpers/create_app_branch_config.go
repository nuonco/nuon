package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

func (h *Helpers) CreateAppBranchConfig(
	ctx context.Context,
	appBranchID string,
	connectedGithubVCSConfig *app.ConnectedGithubVCSConfig,
	publicGitVCSConfig *app.PublicGitVCSConfig,
	installGroups []app.AppBranchInstallGroup,
) (*app.AppBranchConfig, error) {
	config := app.AppBranchConfig{
		AppBranchID:   appBranchID,
		InstallGroups: installGroups,
	}

	// Create config first to get ID
	if err := h.db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, fmt.Errorf("unable to create app branch config: %w", err)
	}

	// Set ownership on VCS config if connected GitHub config
	if connectedGithubVCSConfig != nil {
		connectedGithubVCSConfig.ComponentConfigID = config.ID
		connectedGithubVCSConfig.ComponentConfigType = plugins.TableName(h.db, app.AppBranchConfig{})

		if err := h.db.WithContext(ctx).Create(connectedGithubVCSConfig).Error; err != nil {
			return nil, fmt.Errorf("unable to create connected github vcs config: %w", err)
		}

		config.ConnectedGithubVCSConfig = connectedGithubVCSConfig
	}

	// Set ownership on VCS config if public git config
	if publicGitVCSConfig != nil {
		publicGitVCSConfig.ComponentConfigID = config.ID
		publicGitVCSConfig.ComponentConfigType = plugins.TableName(h.db, app.AppBranchConfig{})

		if err := h.db.WithContext(ctx).Create(publicGitVCSConfig).Error; err != nil {
			return nil, fmt.Errorf("unable to create public git vcs config: %w", err)
		}

		config.PublicGitVCSConfig = publicGitVCSConfig
	}

	return &config, nil
}
