package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
)

type GetLatestActiveSandboxBuildRequest struct {
	AppConfigID string `json:"app_config_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field AppConfigID
// @start-to-close-timeout 30s
func (a *Activities) GetLatestActiveSandboxBuild(ctx context.Context, req GetLatestActiveSandboxBuildRequest) (*app.AppSandboxBuild, error) {
	var build app.AppSandboxBuild
	res := a.db.WithContext(ctx).
		Where(app.AppSandboxBuild{
			AppConfigID: req.AppConfigID,
			Status:      app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		First(&build)
	if res.Error == nil {
		return &build, nil
	}
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to get latest active sandbox build: %w", res.Error)
	}

	var current app.AppSandboxConfig
	err := a.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("PublicGitVCSConfig").
		Where(app.AppSandboxConfig{AppConfigID: req.AppConfigID}).
		First(&current).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to load sandbox config for app config %s: %w", req.AppConfigID, err)
	}

	var candidates []app.AppSandboxBuild
	err = a.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Where(app.AppSandboxBuild{
			AppID:  current.AppID,
			Status: app.AppSandboxBuildStatusActive,
		}).
		Order("created_at DESC").
		Limit(25).
		Find(&candidates).Error
	if err != nil {
		return nil, fmt.Errorf("unable to list active sandbox builds for app %s: %w", current.AppID, err)
	}
	for i := range candidates {
		if configdiff.SandboxConfigsEqual(candidates[i].AppSandboxConfig, current) {
			return &candidates[i], nil
		}
	}

	return nil, nil
}
