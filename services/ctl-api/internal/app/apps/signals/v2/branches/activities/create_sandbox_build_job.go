package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const sandboxBuildOwnerType = "app_sandbox_builds"

type CreateSandboxBuildJobRequest struct {
	BuildID     string `json:"build_id" validate:"required"`
	RunnerID    string `json:"runner_id" validate:"required"`
	LogStreamID string `json:"log_stream_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateSandboxBuildJob(ctx context.Context, req CreateSandboxBuildJobRequest) (*app.RunnerJob, error) {
	var build app.AppSandboxBuild
	if res := a.db.WithContext(ctx).First(&build, "id = ?", req.BuildID); res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox build: %w", res.Error)
	}

	ctx = cctx.SetOrgIDContext(ctx, build.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, build.CreatedByID)

	job, err := a.runnerHelpers.CreateBuildJob(ctx,
		req.RunnerID,
		sandboxBuildOwnerType,
		build.ID,
		app.RunnerJobTypeSandboxBuild,
		app.RunnerJobOperationTypeBuild,
		req.LogStreamID,
		map[string]string{
			"app_id":               build.AppID,
			"app_config_id":        build.AppConfigID,
			"app_sandbox_build_id": build.ID,
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create sandbox build job: %w", err)
	}

	return job, nil
}
