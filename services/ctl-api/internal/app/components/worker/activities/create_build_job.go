package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

const (
	buildOwnerType string = "component_builds"
)

type CreateBuildJobRequest struct {
	BuildID  string
	RunnerID string
	Op       app.RunnerJobOperationType
	Type     app.RunnerJobType
}

// @temporal-gen activity
// @schedule-to-close-timeout 5s
func (a *Activities) CreateBuildJob(ctx context.Context, req *CreateBuildJobRequest) (*app.RunnerJob, error) {
	bld, err := a.getComponentBuild(ctx, req.BuildID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component build: %w", err)
	}

	ctx = middlewares.SetAccountIDContext(ctx, bld.CreatedByID)
	ctx = middlewares.SetOrgIDContext(ctx, bld.OrgID)

	job, err := a.runnersHelpers.CreateBuildJob(ctx, req.RunnerID, buildOwnerType, bld.ID, req.Type, req.Op)
	if err != nil {
		return nil, fmt.Errorf("unable to create build job: %w", err)
	}

	return job, nil
}
