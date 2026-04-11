package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateComponentBuildRecordRequest struct {
	ComponentID string `validate:"required"`
	OrgID       string `validate:"required"`
	CreatedByID string `validate:"required"`

	// GitRef overrides useLatest when set. Used to pin a build to the branch's specific commit.
	GitRef *string
	// VCSConnectionCommitID is a pre-resolved commit to attach to the build record.
	VCSConnectionCommitID *string
}

// CreateComponentBuildRecord creates a component build record without dispatching to the
// event loop. Used by queue signals that handle build execution themselves.
//
// @temporal-gen-v2 activity
func (a *Activities) CreateComponentBuildRecord(ctx context.Context, req CreateComponentBuildRecordRequest) (*app.ComponentBuild, error) {
	ctx = cctx.SetOrgIDContext(ctx, req.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, req.CreatedByID)

	useLatest := req.GitRef == nil
	build, err := a.helpers.CreateComponentBuild(ctx, req.ComponentID, useLatest, req.GitRef)
	if err != nil {
		return nil, fmt.Errorf("create component build: %w", err)
	}

	// If a pre-resolved commit ID was provided and the build doesn't already have one, attach it.
	if req.VCSConnectionCommitID != nil && build.VCSConnectionCommitID == nil {
		if res := a.db.WithContext(ctx).Model(build).Update("vcs_connection_commit_id", *req.VCSConnectionCommitID); res.Error != nil {
			return nil, fmt.Errorf("update build commit: %w", res.Error)
		}
		build.VCSConnectionCommitID = req.VCSConnectionCommitID
	}

	return build, nil
}
