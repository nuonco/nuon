package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateActionWorkflowRunRunnerJob struct {
	ActionWorkflowRunID string            `validate:"required"`
	RunnerID            string            `validate:"required"`
	LogStreamID         string            `validate:"required"`
	Metadata            map[string]string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field ActionWorkflowRunID
func (a *Activities) CreateActionWorkflowRunRunnerJob(ctx context.Context, req *CreateActionWorkflowRunRunnerJob) (*app.RunnerJob, error) {
	run, err := a.getInstallActionWorkflowRun(ctx, req.ActionWorkflowRunID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflow run")
	}

	// adhoc runs have no ActionWorkflowConfig; their timeout lives on the run
	cfg := run.ActionWorkflowConfig
	if cfg.Timeout == 0 {
		cfg.Timeout = run.Timeout
	}

	job, err := a.runnersHelpers.CreateActionsWorkflowRunJob(ctx,
		req.RunnerID,
		req.ActionWorkflowRunID,
		req.LogStreamID,
		&cfg,
		req.Metadata,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create runner job")
	}

	return job, nil
}
