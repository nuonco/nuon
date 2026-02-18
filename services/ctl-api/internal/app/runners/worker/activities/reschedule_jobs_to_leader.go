package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
)

type RescheduleJobsToLeaderRequest struct {
	OldLeaderRunnerID string `validate:"required"`
	NewLeaderRunnerID string `validate:"required"`
}

type RescheduleJobsToLeaderResponse struct {
	RescheduledCount int
}

// @temporal-gen activity
// @by-id NewLeaderRunnerID
func (a *Activities) RescheduleJobsToLeader(ctx context.Context, req RescheduleJobsToLeaderRequest) (*RescheduleJobsToLeaderResponse, error) {
	// Batch-update all queued jobs from old leader to new leader in one query.
	res := a.db.WithContext(ctx).
		Model(&app.RunnerJob{}).
		Where("runner_id = ? AND status = ? AND deleted_at = 0", req.OldLeaderRunnerID, app.RunnerJobStatusQueued).
		Update("runner_id", req.NewLeaderRunnerID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to reschedule jobs to new leader: %w", res.Error)
	}

	// Always signal queued jobs on the new leader to ensure idempotency on retry.
	// Redundant signals are safe; missing signals leave jobs stuck.
	var jobs []app.RunnerJob
	if err := a.db.WithContext(ctx).
		Select("id").
		Where("runner_id = ? AND status = ? AND deleted_at = 0", req.NewLeaderRunnerID, app.RunnerJobStatusQueued).
		Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("unable to fetch queued jobs on new leader: %w", err)
	}

	for _, job := range jobs {
		a.evClient.Send(ctx, req.NewLeaderRunnerID, &signals.Signal{
			Type:  signals.OperationProcessJob,
			JobID: job.ID,
		})
	}

	return &RescheduleJobsToLeaderResponse{RescheduledCount: int(res.RowsAffected)}, nil
}
