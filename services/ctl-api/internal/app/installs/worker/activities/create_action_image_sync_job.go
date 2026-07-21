package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateActionImageSyncJobRequest struct {
	ActionWorkflowRunID string            `validate:"required"`
	RunnerID            string            `validate:"required"`
	LogStreamID         string            `validate:"required"`
	Metadata            map[string]string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ActionWorkflowRunID
func (a *Activities) CreateActionImageSyncJob(ctx context.Context, req *CreateActionImageSyncJobRequest) (*app.RunnerJob, error) {
	run, err := a.getInstallActionWorkflowRun(ctx, req.ActionWorkflowRunID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflow run")
	}

	ctx = cctx.SetAccountIDContext(ctx, run.CreatedByID)
	ctx = cctx.SetOrgIDContext(ctx, run.OrgID)

	job, err := a.runnersHelpers.CreateActionImageSyncJob(ctx,
		req.RunnerID,
		req.ActionWorkflowRunID,
		req.LogStreamID,
		req.Metadata,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create action image sync job")
	}

	return job, nil
}
