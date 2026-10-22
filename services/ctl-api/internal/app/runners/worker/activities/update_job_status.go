package activities

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type UpdateJobStatusRequest struct {
	JobID             string              `validate:"required"`
	Status            app.RunnerJobStatus `validate:"required"`
	StatusDescription string              `validate:"required"`
}

// @temporal-gen-v2 activity
// @max-retries 2
func (a *Activities) UpdateJobStatus(ctx context.Context, req UpdateJobStatusRequest) error {
	runner := app.RunnerJob{
		ID: req.JobID,
	}
	res := a.db.WithContext(ctx).Model(&runner).Updates(app.RunnerJob{
		Status:            req.Status,
		StatusDescription: req.StatusDescription,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update job status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(
			gorm.ErrRecordNotFound,
			fmt.Sprintf("no job found: %s", req.JobID),
		)
	}
	if !req.Status.IsTerminal() {
		return nil
	}

	var job app.RunnerJob
	if err := a.db.WithContext(ctx).Select("id", "runner_id").Where(app.RunnerJob{ID: req.JobID}).Take(&job).Error; err != nil {
		return fmt.Errorf("load job for completion: %w", err)
	}
	return a.enqueueJobCompleted(ctx, job.RunnerID, job.ID)
}

type jobCompletedEnqueue struct {
	JobID    string `json:"job_id"`
	RunnerID string `json:"runner_id"`
}

func (s jobCompletedEnqueue) Type() signal.SignalType { return "job-completed" }

func (s jobCompletedEnqueue) Validate(workflow.Context) error { return nil }

func (s jobCompletedEnqueue) Execute(workflow.Context) error {
	return fmt.Errorf("job-completed is executed from the registered signal")
}

func (a *Activities) enqueueJobCompleted(ctx context.Context, runnerID, jobID string) error {
	if a.queueClient == nil || runnerID == "" {
		return nil
	}
	q, err := a.queueClient.GetQueueByOwnerAndName(ctx, runnerID, "runners", queuenames.RunnerSignalsQueueName)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			return nil
		}
		return fmt.Errorf("find runner-signals queue: %w", err)
	}
	if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:        q.ID,
		OwnerID:        runnerID,
		OwnerType:      plugins.TableName(a.db, app.Runner{}),
		Signal:         jobCompletedEnqueue{JobID: jobID, RunnerID: runnerID},
		IdempotencyKey: "job-completed:" + jobID,
	}); err != nil {
		return fmt.Errorf("enqueue job-completed: %w", err)
	}
	return nil
}
