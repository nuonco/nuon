package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateDeployJobRequest struct {
	RunnerID    string                     `validate:"required"`
	DeployID    string                     `validate:"required"`
	Op          app.RunnerJobOperationType `validate:"required"`
	Type        app.RunnerJobType          `validate:"required"`
	LogStreamID string                     `validate:"required"`
	Metadata    map[string]string          `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateDeployJob(ctx context.Context, req *CreateDeployJobRequest) (*app.RunnerJob, error) {
	deploy, err := a.getDeploy(ctx, req.DeployID)
	if err != nil {
		return nil, fmt.Errorf("unable to get deploy: %w", err)
	}

	ctx = cctx.SetAccountIDContext(ctx, deploy.CreatedByID)
	ctx = cctx.SetOrgIDContext(ctx, deploy.OrgID)

	job, err := a.runnersHelpers.CreateDeployJob(ctx,
		req.RunnerID,
		req.Type,
		req.Op,
		req.DeployID,
		req.LogStreamID,
		req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to create deploy job: %w", err)
	}

	return job, nil
}
