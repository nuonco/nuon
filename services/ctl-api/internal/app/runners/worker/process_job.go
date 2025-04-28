package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ProcessJob(ctx workflow.Context, sreq signals.RequestSignal) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("fetching runner to ensure it is healthy")
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: sreq.ID,
	})
	if err != nil {
		w.updateJobStatus(ctx, sreq.JobID, app.RunnerJobStatusNotAttempted, "unable to fetch runner from database")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	if !runner.Status.IsHealthy() {
		l.Warn("runner is not healthy, not attempting")
		w.updateJobStatus(ctx, sreq.JobID, app.RunnerJobStatusNotAttempted, "runner is not in a healthy state")
		return errors.New("runner is not healthy")
	}

	runnerJob, err := activities.AwaitGetJob(ctx, activities.GetJobRequest{
		ID: sreq.JobID,
	})
	if err != nil {
		w.updateJobStatus(ctx, sreq.JobID, app.RunnerJobStatusNotAttempted, "unable to get job from database")
		return fmt.Errorf("unable to update runner job: %w", err)
	}

	// clear any old jobs behind this that were orphaned, or not attempted for whatever reason. This is usually due
	// to an internal error where something was dropped by temporal. Ideally this would _never_ happen, but it's
	// worth the extra step just in case.
	if err := activities.AwaitFlushOrphanedJobs(ctx, activities.FlushOrphanedJobsRequest{
		RunnerID:  sreq.ID,
		Threshold: runnerJob.CreatedAt.Add(-time.Minute * 5),
	}); err != nil {
		return errors.Wrap(err, "unable to flush orphaned jobs")
	}

	activities.AwaitUpdateJobStartedAt(ctx, activities.UpdateJobStartedAtRequest{JobID: sreq.JobID, StartedAt: workflow.Now(ctx)})
	defer func() {
		// NOTE(fd): wrapped this in a func so the time is set correctly
		activities.AwaitUpdateJobFinishedAt(ctx, activities.UpdateJobFinishedAtRequest{JobID: sreq.JobID, FinishedAt: workflow.Now(ctx)})
	}()

	// Handle sandbox mode, which by _default_ will mimic processing by simply sleeping for 5 seconds (or whatever
	// is currently configured).
	//
	// However, in local development, it can be useful to run a runner and see jobs go all the way through, for
	// testing which can be enabled by changing the sandbox-mode enable runners field.
	if runner.RunnerGroup.Settings.SandboxMode {
		l.Info("runner is in sandbox mode")
		// if the org is a sandbox mode + canary, we do not require the runner locally
		if runner.Org.CreatedBy.AccountType == app.AccountTypeCanary {
			w.updateJobStatus(ctx, sreq.JobID, app.RunnerJobStatusFinished, "success in sandbox/canary mode")
			return nil
		}

		// if sandbox runners are enabled, then force this to run
		if w.cfg.SandboxModeEnableRunners {
			l.Info("enable runners enabled, so this job must be handled by a local runner",
				zap.String("job_id", sreq.JobID),
				zap.String("runner_id", sreq.ID),
			)
		} else {
			// final case is where sandbox runners are disabled
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
		l.Info(fmt.Sprintf("attempting job execution %d of %d", i+1, runnerJob.MaxExecutions))
		retry, started, err := w.startJobExecution(ctx, runnerJob)
		if err != nil {
			return err
		}
		// if the job was not started, we _only_ continue if this was a retryable state
		if !started {
			if !retry {
				return nil
			}
			continue
		}

		// job was started, and the execution
		retry, err = w.monitorJobExecution(ctx, runnerJob)
		if err != nil {
			return err
		}
		if !retry {
			return nil
		}
	}

	return nil
}
