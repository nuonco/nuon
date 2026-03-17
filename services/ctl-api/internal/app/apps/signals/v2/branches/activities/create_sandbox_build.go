package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateSandboxBuildRequest struct {
	AppID              string `json:"app_id" validate:"required"`
	AppConfigID        string `json:"app_config_id" validate:"required"`
	AppSandboxConfigID string `json:"app_sandbox_config_id" validate:"required"`
	OrgID              string `json:"org_id" validate:"required"`
	CreatedByID        string `json:"created_by_id" validate:"required"`

	VCSConnectionCommitID *string `json:"vcs_connection_commit_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateSandboxBuild(ctx context.Context, req CreateSandboxBuildRequest) (*app.AppSandboxBuild, error) {
	ctx = cctx.SetOrgIDContext(ctx, req.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, req.CreatedByID)

	build := app.AppSandboxBuild{
		AppID:              req.AppID,
		AppConfigID:        req.AppConfigID,
		AppSandboxConfigID: req.AppSandboxConfigID,
		Status:             app.AppSandboxBuildStatusQueued,
		StatusDescription:  "queued and waiting for runner",
	}
	if req.VCSConnectionCommitID != nil {
		build.VCSConnectionCommitID = req.VCSConnectionCommitID
	}

	if res := a.db.WithContext(ctx).Create(&build); res.Error != nil {
		return nil, fmt.Errorf("unable to create sandbox build: %w", res.Error)
	}

	return &build, nil
}
