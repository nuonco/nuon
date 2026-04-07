package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type GetRunnerJobGroupQueueRequest struct {
	RunnerID string `validate:"required"`
	JobID    string `validate:"required"`
}

type GetRunnerJobGroupQueueResponse struct {
	// QueueID is the ID of the job-group queue to use, or empty string if the
	// parallel-runner-jobs feature is not enabled or no queue exists for the group.
	QueueID string
}

// GetRunnerJobGroupQueue returns the per-job-group queue ID for the given runner+job pair
// when the parallel-runner-jobs feature flag is enabled, or empty string otherwise.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) PkgWorkflowsJobGetRunnerJobGroupQueue(ctx context.Context, req *GetRunnerJobGroupQueueRequest) (*GetRunnerJobGroupQueueResponse, error) {
	runner := app.Runner{}
	if res := a.db.WithContext(ctx).
		Preload("Org").
		Preload("Queues").
		First(&runner, "id = ?", req.RunnerID); res.Error != nil {
		return &GetRunnerJobGroupQueueResponse{}, nil
	}

	if !runner.Org.Features[string(app.OrgFeatureParallelRunnerJobs)] {
		return &GetRunnerJobGroupQueueResponse{}, nil
	}

	job := app.RunnerJob{}
	if res := a.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		First(&job, "id = ?", req.JobID); res.Error != nil {
		return &GetRunnerJobGroupQueueResponse{}, nil
	}

	q := runner.GetQueueForGroup(job.Type.Group())
	if q == nil {
		return &GetRunnerJobGroupQueueResponse{}, nil
	}

	return &GetRunnerJobGroupQueueResponse{QueueID: q.ID}, nil
}
