package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

type GetInstallActionWorkflowStateRequest struct {
	InstallActionWorkflowID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallActionWorkflowID
func (a *Activities) GetInstallActionWorkflowState(ctx context.Context, req GetInstallActionWorkflowStateRequest) (*app.InstallActionWorkflow, error) {
	var act app.InstallActionWorkflow
	res := a.db.WithContext(ctx).
		Preload("ActionWorkflow").
		Find(&act, "id= ?", req.InstallActionWorkflowID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install action workflow state")
	}

	runsByWorkflow, err := a.helpers.GetLatestInstallActionWorkflowRuns(ctx, []string{act.ID}, helpers.LatestInstallActionWorkflowRunsOptions{WithRunnerJobs: true})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest install action workflow runs")
	}
	act.Runs = runsByWorkflow[act.ID]

	return &act, nil
}
