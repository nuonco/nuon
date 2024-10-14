package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateSyncJobRequest struct {
	RunnerID string
	DeployID string
	Op       app.RunnerJobOperationType
	Type     app.RunnerJobType
}

// @temporal-gen activity
func (a *Activities) CreateSyncJob(ctx context.Context, req *CreateSyncJobRequest) (*app.RunnerJob, error) {
	deploy, err := a.getDeploy(ctx, req.DeployID)
	if err != nil {
		return nil, fmt.Errorf("unable to get deploy: %w", err)
	}

	ctx = cctx.SetAccountIDContext(ctx, deploy.CreatedByID)
	ctx = cctx.SetOrgIDContext(ctx, deploy.OrgID)

	job, err := a.runnersHelpers.CreateSyncJob(ctx, req.RunnerID, req.Type, req.Op, req.DeployID)
	if err != nil {
		return nil, fmt.Errorf("unable to create install sandbox job: %w", err)
	}

	return job, nil
}
