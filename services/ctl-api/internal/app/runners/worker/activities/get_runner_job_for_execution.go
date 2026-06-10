package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetRunnerJobForExecutionRequest struct {
	RunnerID string `validate:"required"`
	JobID    string `validate:"required"`
}

type GetRunnerJobForExecutionResponse struct {
	Runner           *app.Runner
	Job              *app.RunnerJob
	HasActiveProcess bool
}

// GetRunnerJobForExecution coalesces the three reads the process_job execute
// phase needs — runner, active-process check, and job — into a single activity.
// Done separately these were three serial Temporal round-trips on the dispatch
// hot path right before started_at is stamped; folding them into one round-trip
// shaves that latency for every job.
//
// @temporal-gen-v2 activity
func (a *Activities) GetRunnerJobForExecution(ctx context.Context, req GetRunnerJobForExecutionRequest) (*GetRunnerJobForExecutionResponse, error) {
	runner, err := a.getRunner(ctx, req.RunnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner: %w", err)
	}

	var activeCount int64
	if err := a.db.WithContext(ctx).
		Model(&app.RunnerProcess{}).
		Where("runner_id = ? AND composite_status->>'status' = ?", req.RunnerID, string(app.RunnerProcessStatusActive)).
		Count(&activeCount).Error; err != nil {
		return nil, fmt.Errorf("unable to check active runner process: %w", err)
	}

	job, err := a.getRunnerJob(ctx, req.JobID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner job: %w", err)
	}

	return &GetRunnerJobForExecutionResponse{
		Runner:           runner,
		Job:              job,
		HasActiveProcess: activeCount > 0,
	}, nil
}
