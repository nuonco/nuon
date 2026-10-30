package helpers

import (
	"context"

	"github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type LatestInstallActionWorkflowRunsOptions struct {
	WithRunnerJobs  bool
	RunnerJobSelect []string
}

// GetLatestInstallActionWorkflowRuns returns at most one latest non-deleted run per install action workflow ID, keyed
// by workflow ID. Unlike install_action_workflow_runs_state_view_v1, it does not read each workflow's full run history.
func (h *Helpers) GetLatestInstallActionWorkflowRuns(ctx context.Context, workflowIDs []string, opts LatestInstallActionWorkflowRunsOptions) (map[string][]app.InstallActionWorkflowRun, error) {
	byWorkflow := make(map[string][]app.InstallActionWorkflowRun, len(workflowIDs))
	if len(workflowIDs) == 0 {
		return byWorkflow, nil
	}

	var runs []app.InstallActionWorkflowRun
	if err := h.db.WithContext(ctx).Raw(`
		SELECT r.*
		FROM unnest(?::text[]) AS w(id)
		CROSS JOIN LATERAL (
			SELECT *
			FROM install_action_workflow_runs r
			WHERE r.install_action_workflow_id = w.id AND r.deleted_at = 0
			ORDER BY r.created_at DESC
			LIMIT 1
		) r
	`, pq.Array(workflowIDs)).Find(&runs).Error; err != nil {
		return nil, errors.Wrap(err, "unable to load latest install action workflow runs")
	}

	if opts.WithRunnerJobs && len(runs) > 0 {
		if err := h.attachRunnerJobs(ctx, runs, opts.RunnerJobSelect); err != nil {
			return nil, err
		}
	}

	for _, r := range runs {
		byWorkflow[r.InstallActionWorkflowID.String] = append(byWorkflow[r.InstallActionWorkflowID.String], r)
	}

	return byWorkflow, nil
}

func (h *Helpers) attachRunnerJobs(ctx context.Context, runs []app.InstallActionWorkflowRun, selectCols []string) error {
	runIDs := make([]string, 0, len(runs))
	for _, r := range runs {
		runIDs = append(runIDs, r.ID)
	}

	var jobs []app.RunnerJob
	query := h.db.WithContext(ctx).
		Where(app.RunnerJob{OwnerType: "install_action_workflow_runs"}).
		Where("owner_id IN ?", runIDs)
	if len(selectCols) > 0 {
		query = query.Select(selectCols)
	}
	if err := query.Find(&jobs).Error; err != nil {
		return errors.Wrap(err, "unable to load runner jobs for install action workflow runs")
	}

	jobsByOwnerID := make(map[string]*app.RunnerJob, len(jobs))
	for i := range jobs {
		jobsByOwnerID[jobs[i].OwnerID] = &jobs[i]
	}

	for i := range runs {
		job, ok := jobsByOwnerID[runs[i].ID]
		if !ok {
			continue
		}
		runs[i].RunnerJob = job
		if err := runs[i].AfterQuery(h.db); err != nil {
			return errors.Wrap(err, "unable to hydrate install action workflow run")
		}
	}

	return nil
}
