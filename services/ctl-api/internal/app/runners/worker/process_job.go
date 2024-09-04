package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

func (w *Workflows) processJob(ctx workflow.Context, runnerID, jobID string) error {
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateJobStatus(ctx, jobID, app.RunnerJobStatusNotAttempted, "unable to fetch runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	if !runner.Status.IsHealthy() {
		w.updateJobStatus(ctx, jobID, app.RunnerJobStatusNotAttempted, "runner is not in a healthy state")
		return nil
	}

	runnerJob, err := activities.AwaitGetRunnerJob(ctx, jobID)
	if err != nil {
		w.updateJobStatus(ctx, jobID, app.RunnerJobStatusNotAttempted, "unable to get job from database")
		return fmt.Errorf("unable to get runner job: %w", err)
	}

	now := workflow.Now(ctx)
	if runnerJob.CreatedAt.Add(runnerJob.QueueTimeout).Before(now) {
		w.updateJobStatus(ctx, jobID, app.RunnerJobStatusNotAttempted, "queue timeout reached")
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
