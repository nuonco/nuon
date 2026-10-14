package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
	if run.InstallWorkflowID != nil {
		ctx = cctx.SetFlowWorkflowIDContext(ctx, *run.InstallWorkflowID)
	}
	ctx = cctx.SetFlowInstallIDContext(ctx, run.InstallID)

	group := actionRunnerJobGroup(run.TriggerType, cfg.Image)
	if group == app.RunnerJobGroupCronActions {
		enabled, err := a.ensureRunnerJobGroup(ctx, req.RunnerID, group)
		if err != nil {
			return nil, err
		}
		if !enabled {
			group = app.RunnerJobGroupActions
		}
	}

	job, err := a.runnersHelpers.CreateActionsWorkflowRunJob(ctx,
		req.RunnerID,
		req.ActionWorkflowRunID,
		req.LogStreamID,
		&cfg,
		group,
		req.Metadata,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create runner job")
	}

	return job, nil
}

func actionRunnerJobGroup(trigger app.ActionWorkflowTriggerType, image string) app.RunnerJobGroup {
	if image != "" {
		return app.RunnerJobGroupImageActions
	}
	if trigger == app.ActionWorkflowTriggerTypeCron {
		return app.RunnerJobGroupCronActions
	}
	return app.RunnerJobGroupActions
}

func (a *Activities) ensureRunnerJobGroup(ctx context.Context, runnerID string, group app.RunnerJobGroup) (bool, error) {
	var runner app.Runner
	if err := a.db.WithContext(ctx).
		Preload("Queues").
		Preload("RunnerGroup.Settings").
		First(&runner, "id = ?", runnerID).Error; err != nil {
		return false, errors.Wrap(err, "unable to get runner")
	}
	if runner.GetQueueForGroup(group) != nil {
		return true, nil
	}
	if runner.GetQueueForGroup(app.RunnerJobGroupActions) == nil {
		return false, nil
	}
	if err := a.runnersHelpers.EnsureRunnerJobGroupQueue(ctx, &runner, &runner.RunnerGroup.Settings, group); err != nil {
		return false, err
	}
	return true, nil
}
