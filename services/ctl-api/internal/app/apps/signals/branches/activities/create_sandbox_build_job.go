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
	RunnerID    string `json:"runner_id,omitempty"`
	LogStreamID string `json:"log_stream_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateSandboxBuildJob(ctx context.Context, req CreateSandboxBuildJobRequest) (*app.RunnerJob, error) {
	var build app.AppSandboxBuild
	if res := a.db.WithContext(ctx).Where(&app.AppSandboxBuild{ID: req.BuildID}).First(&build); res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox build: %w", res.Error)
	}

	ctx = cctx.SetOrgIDContext(ctx, build.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, build.CreatedByID)
	executor, runnerID, err := a.runnerHelpers.BuildExecutorForOrg(ctx, &app.Org{ID: build.OrgID}, app.RunnerJobTypeSandboxBuild)
	if err != nil {
		return nil, fmt.Errorf("unable to choose build executor: %w", err)
	}
	if executor == app.RunnerJobExecutorOrgRunner && req.RunnerID != "" {
		runnerID = req.RunnerID
	}

	job, err := a.runnerHelpers.CreateBuildJob(ctx,
		runnerID,
		executor,
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
