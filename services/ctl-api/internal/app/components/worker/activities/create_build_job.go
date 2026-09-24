package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	buildOwnerType string = "component_builds"
)

type CreateBuildJobRequest struct {
	BuildID     string                     `validate:"required"`
	Op          app.RunnerJobOperationType `validate:"required"`
	Type        app.RunnerJobType          `validate:"required"`
	LogStreamID string                     `validate:"required"`
	Metadata    map[string]string          `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateBuildJob(ctx context.Context, req *CreateBuildJobRequest) (*app.RunnerJob, error) {
	bld, err := a.getComponentBuildWithConfig(ctx, req.BuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component build: %w", err)
	}

	if req.Type == app.RunnerJobTypeDockerBuild {
		return nil, fmt.Errorf("docker_build components are not supported by control-plane builds; replace this component with a container_image component")
	}

	ctx = cctx.SetAccountIDContext(ctx, bld.CreatedByID)
	ctx = cctx.SetOrgIDContext(ctx, bld.OrgID)

	job, err := a.runnersHelpers.CreateBuildJob(ctx,
		buildOwnerType,
		bld.ID,
		req.Type,
		req.Op,
		req.LogStreamID,
		req.Metadata,
		bld.ComponentConfigConnection.GetBuildTimeout(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create build job: %w", err)
	}

	return job, nil
}
