package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get latest active sandbox build: %w", res.Error)
	}

	return &build, nil
}
