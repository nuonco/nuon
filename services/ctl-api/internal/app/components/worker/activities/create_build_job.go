package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

const (
	buildOwnerType string = "component_builds"
)

type CreateBuildJobRequest struct {
	BuildID     string                     `validate:"required"`
	RunnerID    string                     `validate:"required"`
	Op          app.RunnerJobOperationType `validate:"required"`
	Type        app.RunnerJobType          `validate:"required"`
	LogStreamID string                     `validate:"required"`
	Metadata    map[string]string          `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateBuildJob(ctx context.Context, req *CreateBuildJobRequest) (*app.RunnerJob, error) {
	bld, err := a.getComponentBuild(ctx, req.BuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component build: %w", err)
	}

	ctx = cctx.SetAccountIDContext(ctx, bld.CreatedByID)
	ctx = cctx.SetOrgIDContext(ctx, bld.OrgID)

	job, err := a.runnersHelpers.CreateBuildJob(ctx,
		req.RunnerID,
		buildOwnerType,
		bld.ID,
		req.Type,
		req.Op,
		req.LogStreamID,
		req.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create build job: %w", err)
	}

	return job, nil
}
