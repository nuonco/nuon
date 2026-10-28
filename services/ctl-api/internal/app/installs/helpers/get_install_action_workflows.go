package helpers

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// getInstallActionWorkflows reads action workflows DB.
func (h *Helpers) getInstallActionWorkflows(ctx context.Context, installID string, withRunnerJobOutputs bool) ([]app.InstallActionWorkflow, error) {
	var acts []app.InstallActionWorkflow
	res := h.db.WithContext(ctx).
		Preload("ActionWorkflow").
		Find(&acts, "install_id = ?", installID)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install components")
	}

	workflowIDs := make([]string, 0, len(acts))
	for _, act := range acts {
		workflowIDs = append(workflowIDs, act.ID)
	}

	opts := LatestInstallActionWorkflowRunsOptions{}
	if withRunnerJobOutputs {
		opts.WithRunnerJobs = true
		opts.RunnerJobSelect = []string{"id", "owner_id", "owner_type", "type", "outputs"}
	}
	runsByWorkflow, err := h.GetLatestInstallActionWorkflowRuns(ctx, workflowIDs, opts)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest install action workflow runs")
	}
	for i := range acts {
		acts[i].Runs = runsByWorkflow[acts[i].ID]
	}

	return acts, nil
}
