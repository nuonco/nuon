package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateRunnerJobRequest struct {
	InstallActionWorkflowRunID string
	RunnerID                   string
}

// @temporal-gen activity
func (a *Activities) CreateRunnerJob(ctx context.Context, req *CreateRunnerJobRequest) (*app.RunnerJob, error) {
	run, err := a.getInstallActionWorkflowRun(ctx, req.InstallActionWorkflowRunID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get deploy")
	}

	job, err := a.runnersHelpers.CreateActionsWorkflowRunJob(ctx,
		req.RunnerID,
		req.InstallActionWorkflowRunID,
		run.LogStream.ID,
		&run.ActionWorkflowConfig,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create runner job")
	}

	return job, nil
}
