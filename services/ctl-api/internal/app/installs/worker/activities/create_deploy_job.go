package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

type CreateDeployJobRequest struct {
	RunnerID string
	DeployID string
	Op       app.RunnerJobOperationType
	Type     app.RunnerJobType
}

// @temporal-gen activity
func (a *Activities) CreateDeployJob(ctx context.Context, req *CreateDeployJobRequest) (*app.RunnerJob, error) {
	deploy, err := a.getDeploy(ctx, req.DeployID)
	if err != nil {
		return nil, fmt.Errorf("unable to get deploy: %w", err)
	}

	ctx = middlewares.SetAccountIDContext(ctx, deploy.CreatedByID)
	ctx = middlewares.SetOrgIDContext(ctx, deploy.OrgID)

	job, err := a.runnersHelpers.CreateDeployJob(ctx, req.RunnerID, req.Type, req.Op, req.DeployID)
	if err != nil {
		return nil, fmt.Errorf("unable to create deploy job: %w", err)
	}

	return job, nil
}
