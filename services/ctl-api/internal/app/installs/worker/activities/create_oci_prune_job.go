package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateOCIPruneJobRequest struct {
	RunnerID    string            `validate:"required"`
	InstallID   string            `validate:"required"`
	LogStreamID string            `validate:"required"`
	Metadata    map[string]string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateOCIPruneJob(ctx context.Context, req *CreateOCIPruneJobRequest) (*app.RunnerJob, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	ctx = cctx.SetAccountIDContext(ctx, install.CreatedByID)
	ctx = cctx.SetOrgIDContext(ctx, install.OrgID)

	job, err := a.runnersHelpers.CreateOCIPruneJob(ctx, req.RunnerID, req.InstallID, req.LogStreamID, req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to create oci prune job: %w", err)
	}
	return job, nil
}
