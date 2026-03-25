package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type QueueComponentBuildRequest struct {
	ComponentID string `validate:"required"`
	OrgID       string `validate:"required"`
	CreatedByID string `validate:"required"`

	// GitRef overrides useLatest when set. Used to pin a build to the branch's specific commit.
	GitRef *string
	// VCSConnectionCommitID is a pre-resolved commit to attach to the build record.
	VCSConnectionCommitID *string
}

// @temporal-gen-v2 activity
func (a *Activities) QueueComponentBuild(ctx context.Context, req QueueComponentBuildRequest) (*app.ComponentBuild, error) {
	// set the orgID on the context, for all writes
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

	// Ensure the component event loop is running before sending the build signal.
	// In the normal flow, the event loop is started by OperationCreated when the component
	// is created via the HTTP API. But in the onboarding flow (and other paths that create
	// components through SyncAppConfig), the event loop may not exist yet.
	// OperationRestart has Restart()=true, so evClient.Send checks the workflow status
	// and starts the event loop if it's not running. The Restarted handler is a no-op
	// (just sets component status to active).
	a.evClient.Send(ctx, req.ComponentID, &signals.Signal{
		Type: signals.OperationRestart,
	})

	a.evClient.Send(ctx, req.ComponentID, &signals.Signal{
		Type:    signals.OperationBuild,
		BuildID: build.ID,
	})
	return build, nil
}
