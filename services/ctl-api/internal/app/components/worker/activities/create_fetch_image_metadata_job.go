package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateFetchImageMetadataJobRequest struct {
	BuildID     string            `validate:"required"`
	RunnerID    string            `validate:"required"`
	LogStreamID string            `validate:"required"`
	Metadata    map[string]string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateFetchImageMetadataJob(ctx context.Context, req *CreateFetchImageMetadataJobRequest) (*app.RunnerJob, error) {
	bld, err := a.getComponentBuild(ctx, req.BuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component build: %w", err)
	}

	ctx = cctx.SetAccountIDContext(ctx, bld.CreatedByID)
	ctx = cctx.SetOrgIDContext(ctx, bld.OrgID)

	job, err := a.runnersHelpers.CreateFetchImageMetadataJob(ctx,
		req.RunnerID,
		buildOwnerType,
		bld.ID,
		req.LogStreamID,
		req.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create fetch image metadata job: %w", err)
	}

	return job, nil
}
