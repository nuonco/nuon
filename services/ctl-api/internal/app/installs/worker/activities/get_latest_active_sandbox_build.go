package activities

import (
	"context"

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
	return configdiff.LatestActiveSandboxBuild(ctx, a.db, req.AppConfigID)
}
