package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) CreateAppBranchConfig(
	ctx context.Context,
	appBranchID string,
	connectedGithubVCSConfig *app.ConnectedGithubVCSConfig,
	publicGitVCSConfig *app.PublicGitVCSConfig,
	installGroups []app.AppBranchInstallGroup,
) (*app.AppBranchConfig, error) {
	config := app.AppBranchConfig{
		AppBranchID:              appBranchID,
		InstallGroups:            installGroups,
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,
		PublicGitVCSConfig:       publicGitVCSConfig,
	}

	var previous app.AppBranchConfig
	res := h.db.WithContext(ctx).
		Where(app.AppBranchConfig{AppBranchID: appBranchID}).
		Order("created_at DESC").
		First(&previous)
	if res.Error == nil {
		config.ComponentIDs = previous.ComponentIDs
		config.ActionIDs = previous.ActionIDs
		config.RunbookIDs = previous.RunbookIDs
	} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to load previous app branch config: %w", res.Error)
	}

	if err := h.db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, fmt.Errorf("unable to create app branch config: %w", err)
	}

	return &config, nil
}
