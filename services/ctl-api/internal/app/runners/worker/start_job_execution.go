package worker

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

const (
	defaultJobPollPeriod       time.Duration = time.Second * 1
	defaultAvailablePollPeriod time.Duration = time.Second * 1
)

// this function is the most core part of the runner job system, it's responsible for a.) marking a job as available and
// then b.) waiting until it is picked up by a runner (ie: an execution exists) and then c.) finished.
//
// it is responsible for updating the runner job with each state, and in some cases the runner job execution.
func (w *Workflows) startJobExecution(ctx workflow.Context, job *app.RunnerJob) (bool, bool, error) {
	startTS := workflow.Now(ctx)
	tags := map[string]string{
		"status":   "ok",
		"job_type": string(job.Type),
		"job_group": string(job.Group),
	}
	defer func() {
		w.mw.Incr(ctx, "runner.job_execution_start", metrics.ToTags(tags)...)
		e2eLatency := workflow.Now(ctx).Sub(startTS)
		w.mw.Timing(ctx, "runner.job_execution_start_latency", e2eLatency, metrics.ToTags(tags)...)
	}()

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, false, err
	}

	availableStart := workflow.Now(ctx)
	l.Info("marking job as available for runner to pick up")
	w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusAvailable, "waiting for runner to reserve job")

	start := workflow.Now(ctx)
	overallTimeout := job.CreatedAt.Add(job.OverallTimeout)
	availableTimeout := start.Add(job.AvailableTimeout)

	var jobExecutionFound bool
	var runnerStatus app.RunnerStatus
	for runnerStatus != app.RunnerStatusActive {
		workflow.Sleep(ctx, defaultAvailablePollPeriod)

		now := workflow.Now(ctx)
		if now.After(overallTimeout) {
			l.Error("overall job timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout waiting for runner to be healthy")
			tags["status"] = "runner_unhealthy"
			return false, false, nil
		}

		if now.After(availableTimeout) {
			l.Error("timeout waiting for job to be picked up")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "timeout waiting for runner to become healthy")
			tags["status"] = "runner_unhealthy"
			return true, false, nil
		}

		// if the runner is deemed unhealthy, the job execution is marked as unknown, and the job is marked as
		// not attempted with the correct status, this is retryable.
		runnerStatus, err = activities.AwaitGetRunnerStatusByID(ctx, job.RunnerID)
		if err != nil {
			l.Warn("unable to determine runner status", zap.Error(err))
			return false, false, err
		}

		jobStatus, err := activities.AwaitGetJobStatusByID(ctx, job.ID)
		if err != nil {
			return false, false, nil
		}
		if jobStatus == app.RunnerJobStatusCancelled {
			l.Error("job was cancelled")
			tags["status"] = "job_cancelled"
			return false, false, nil
		}
	}

	// poll until the job is picked up, and an execution exists
	for !jobExecutionFound {
		workflow.Sleep(ctx, defaultAvailablePollPeriod)

		now := workflow.Now(ctx)
		if now.After(overallTimeout) {
			l.Error("overall job timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout")
			tags["status"] = "overall_timeout"
			return false, false, nil
		}

		if now.After(availableTimeout) {
			l.Error("timeout waiting for job to be picked up")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "timeout waiting for runner to pick up job")
			tags["status"] = "available_timeout"
			return true, false, nil
		}

		jobStatus, err := activities.AwaitGetJobStatusByID(ctx, job.ID)
		if err != nil {
			return false, false, nil
		}
		if jobStatus == app.RunnerJobStatusCancelled {
			l.Error("job was cancelled")
			tags["status"] = "job_cancelled"
			return false, false, nil
		}

		jobExecutionResp, err := activities.AwaitGetLatestJobExecution(ctx, activities.GetLatestJobExecutionRequest{
			JobID:       job.ID,
			AvailableAt: availableStart,
		})
		if err != nil {
			return false, false, fmt.Errorf("error fetching latest job execution: %w", err)
		}
		jobExecutionFound = jobExecutionResp.Found
	}

	l.Info("job picked up by runner and is in progress")
	return true, true, nil
}
