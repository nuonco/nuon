package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateSandboxBuildRequest struct {
	AppID              string `json:"app_id" validate:"required"`
	AppConfigID        string `json:"app_config_id" validate:"required"`
	AppSandboxConfigID string `json:"app_sandbox_config_id" validate:"required"`
	OrgID              string `json:"org_id" validate:"required"`
	CreatedByID        string `json:"created_by_id" validate:"required"`

	VCSConnectionCommitID *string                  `json:"vcs_connection_commit_id"`
	AppBranchRunID        string                   `json:"app_branch_run_id"`
	SourceIdentity        *sourceidentity.Identity `json:"source_identity,omitempty"`
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
		SourceIdentity:     req.SourceIdentity,
	}
	if req.SourceIdentity != nil {
		if err := req.SourceIdentity.Validate(); err != nil {
			return nil, err
		}
		build.Checksum = req.SourceIdentity.ConfigChecksum
	}
	if req.VCSConnectionCommitID != nil {
		build.VCSConnectionCommitID = req.VCSConnectionCommitID
	}
	if req.AppBranchRunID != "" {
		build.AppBranchRunID = &req.AppBranchRunID
	}

	if err := buildselection.WithMutableConfig(ctx, a.db, req.AppConfigID, func(tx *gorm.DB, cfg *app.AppConfig) error {
		if cfg.BuildsPinned && build.AppID != cfg.AppID {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Create(&build).Error; err != nil {
			return fmt.Errorf("unable to create sandbox build: %w", err)
		}
		if !cfg.BuildsPinned {
			return nil
		}
		return tx.Model(&app.AppConfig{}).Where(app.AppConfig{ID: cfg.ID}).
			Update("sandbox_build_id", generics.NewNullString(build.ID)).Error
	}); err != nil {
		return nil, err
	}

	return &build, nil
}
