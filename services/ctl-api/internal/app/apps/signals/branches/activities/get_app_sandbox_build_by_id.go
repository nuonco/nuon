package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field buildID
func (a *Activities) getAppSandboxBuildByID(ctx context.Context, buildID string) (*app.AppSandboxBuild, error) {
	var build app.AppSandboxBuild
	res := a.db.WithContext(ctx).First(&build, "id = ?", buildID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox build: %w", res.Error)
	}
	return &build, nil
}
