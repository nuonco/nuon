package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ProcessJob(ctx workflow.Context, sreq signals.RequestSignal) error {
	l := workflow.GetLogger(ctx)
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

	// Handle sandbox mode, which by _default_ will mimic processing by simply sleeping for 5 seconds (or whatever
	// is currently configured).
	//
	// However, in local development, it can be useful to run a runner and see jobs go all the way through, for
	// testing which can be enabled by changing the sandbox-mode enable runners field.
	if runner.RunnerGroup.Settings.SandboxMode {
		if w.cfg.SandboxModeEnableRunners {
			l.Info("enable runners enabled, so this job must be handled by a local runner",
				zap.String("job_id", sreq.JobID),
				zap.String("runner_id", sreq.ID),
			)
		} else {
			l.Info("skipping job, and setting to true in sandbox mode",
				zap.String("job_id", sreq.JobID),
				zap.String("runner_id", sreq.ID),
			)
			workflow.Sleep(ctx, w.cfg.SandboxModeSleep)
			w.updateJobStatus(ctx, sreq.JobID, app.RunnerJobStatusFinished, "success in sandbox mode")
			return nil
		}
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
