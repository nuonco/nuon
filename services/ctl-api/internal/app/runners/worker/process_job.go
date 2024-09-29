package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ProcessJob(ctx workflow.Context, sreq signals.RequestSignal) error {
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateJobStatus(ctx, sreq.JobID, app.RunnerJobStatusNotAttempted, "unable to fetch runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	if !runner.Status.IsHealthy() {
		w.updateJobStatus(ctx, sreq.JobID, app.RunnerJobStatusNotAttempted, "runner is not in a healthy state")
		return nil
	}

	runnerJob, err := activities.AwaitGetJob(ctx, activities.GetJobRequest{
		ID: sreq.JobID,
	})
	if err != nil {
		w.updateJobStatus(ctx, sreq.JobID, app.RunnerJobStatusNotAttempted, "unable to get job from database")
		return fmt.Errorf("unable to get runner job: %w", err)
	}

	now := workflow.Now(ctx)
	if runnerJob.CreatedAt.Add(runnerJob.QueueTimeout).Before(now) {
		w.updateJobStatus(ctx, sreq.JobID, app.RunnerJobStatusNotAttempted, "queue timeout reached")
		return nil
	}

	for i := 0; i < runnerJob.MaxExecutions; i++ {
		retry, err := w.processJobExecution(ctx, runnerJob)
		if err != nil {
			return err
		}

		if !retry {
			return nil
		}
	}

	return nil
}
