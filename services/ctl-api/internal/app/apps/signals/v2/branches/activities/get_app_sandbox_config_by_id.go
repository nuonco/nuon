package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field configID
func (a *Activities) getAppSandboxConfigByID(ctx context.Context, configID string) (*app.AppSandboxConfig, error) {
	var cfg app.AppSandboxConfig
	res := a.db.WithContext(ctx).
		Preload("PublicGitVCSConfig").
		Preload("ConnectedGithubVCSConfig").
		First(&cfg, "id = ?", configID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox config: %w", res.Error)
	}
	return &cfg, nil
}
