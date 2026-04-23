package worker

import (
	"fmt"
	"maps"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

// this function is the most core part of the runner job system, it's responsible for a.) marking a job as available and
// then b.) waiting until it is picked up by a runner (ie: an execution exists) and then c.) finished.
//
// it is responsible for updating the runner job with each state, and in some cases the runner job execution.
//
// Waits are push-based: external state writers fire workflow updates (see
// processjob_updates.go) that set flags on processJobState, and this
// function uses workflow.Await to block until a flag or a wall-clock
// deadline fires. Re-reads authoritative state from the DB on wake.
func (w *Workflows) startJobExecution(ctx workflow.Context, state *processJobState, job *app.RunnerJob) (bool, bool, error) {
	startTS := workflow.Now(ctx)
	tags := map[string]string{
		"status":        "ok",
		"job_type":      string(job.Type),
		"job_group":     string(job.Group),
		"job_operation": string(job.Operation),
	}

	etags := maps.Clone(tags)
	etags["job_id"] = job.ID
	etags["job_operation"] = string(job.Operation)
	etags["runner_id"] = job.RunnerID
	etags["org_id"] = string(job.OrgID)
	etags["org_name"] = job.Org.Name
	etags["available_timeout"] = job.AvailableTimeout.String()
	etags["overall_timeout"] = job.OverallTimeout.String()

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

	// Phase 1: wait for runner to be Active (skipped for operations group).
	if job.Group != app.RunnerJobGroupOperations {
		retry, cont, err := w.awaitRunnerActive(ctx, l, state, job, tags, etags, overallTimeout, availableTimeout)
		if err != nil || !cont {
			return retry, false, err
		}
	}

	// Phase 2: wait for the runner to pick up the job (execution exists).
	return w.awaitJobPickedUp(ctx, l, state, job, tags, etags, availableStart, overallTimeout, availableTimeout)
}

// awaitRunnerActive blocks until the runner reports Active status or a
// deadline fires. Returns (retry, continueToPhase2, err). continueToPhase2
// is false if the caller should exit startJobExecution (timeout /
// cancellation).
func (w *Workflows) awaitRunnerActive(
	ctx workflow.Context,
	l *zap.Logger,
	state *processJobState,
	job *app.RunnerJob,
	tags, etags map[string]string,
	overallTimeout, availableTimeout time.Time,
) (bool, bool, error) {
	// Initial-state read covers the race where the runner was already
	// Active before we registered update handlers.
	runnerStatus, err := activities.AwaitGetRunnerStatusByID(ctx, job.RunnerID)
	if err != nil {
		l.Warn("unable to determine runner status", zap.Error(err))
		return false, false, err
	}
	if runnerStatus == app.RunnerStatusActive {
		return false, true, nil
	}

	for {
		now := workflow.Now(ctx)
		if now.After(overallTimeout) {
			l.Error("overall job timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout waiting for runner to be healthy")
			tags["status"] = "runner_unhealthy"
			etags["runner_status"] = string(runnerStatus)
			w.mw.Event(ctx, &statsd.Event{
				Title:          "Overall job timeout reached waiting for runner to become healthy",
				Text:           "Overall end-to-end job execution timeout reached while waiting for job to bewcome healthy",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-timeout-waiting-for-healthy-runner",
			})
			return false, false, nil
		}
		if now.After(availableTimeout) {
			l.Error("timeout waiting for job to be picked up")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "timeout waiting for runner to become healthy")
			tags["status"] = "runner_unhealthy"
			etags["runner_status"] = string(runnerStatus)
			w.mw.Event(ctx, &statsd.Event{
				Title:          "Available timeout reached waiting for runner to become healthy",
				Text:           "Job is ready for execution, but runner did not become healthy within the available timeout",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Low,
				AlertType:      statsd.Warning,
				AggregationKey: "runner-job-timeout-waiting-for-healthy-runner",
			})
			return true, false, nil
		}

		// Wait for a state flag or whichever deadline comes first.
		// AwaitWithTimeout guarantees we wake to re-check the wall-clock
		// deadlines at the top of the loop even if no update ever fires.
		deadline := availableTimeout
		if overallTimeout.Before(deadline) {
			deadline = overallTimeout
		}
		waitDuration := deadline.Sub(now)
		if waitDuration <= 0 {
			continue
		}
		if _, err := workflow.AwaitWithTimeout(ctx, waitDuration, state.anyFlag); err != nil {
			return false, false, err
		}

		if state.jobStatusChanged {
			state.jobStatusChanged = false
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
		if state.runnerStatusChanged {
			state.runnerStatusChanged = false
			runnerStatus, err = activities.AwaitGetRunnerStatusByID(ctx, job.RunnerID)
			if err != nil {
				l.Warn("unable to determine runner status", zap.Error(err))
				return false, false, err
			}
			if runnerStatus == app.RunnerStatusActive {
				return false, true, nil
			}
		}
		// execCreated, execStatusChanged, runnerRestarted here are
		// noise at this phase — just clear them.
		state.execCreated = false
		state.execStatusChanged = false
		state.runnerRestarted = false
	}
}

// awaitJobPickedUp blocks until the runner creates a RunnerJobExecution row
// (job picked up) or a deadline fires.
func (w *Workflows) awaitJobPickedUp(
	ctx workflow.Context,
	l *zap.Logger,
	state *processJobState,
	job *app.RunnerJob,
	tags, etags map[string]string,
	availableStart time.Time,
	overallTimeout, availableTimeout time.Time,
) (bool, bool, error) {
	// Initial-state read for the race.
	execResp, err := activities.AwaitGetLatestJobExecution(ctx, activities.GetLatestJobExecutionRequest{
		JobID:       job.ID,
		AvailableAt: availableStart,
	})
	if err != nil {
		return false, false, fmt.Errorf("error fetching latest job execution: %w", err)
	}
	if execResp.Found {
		l.Info("job picked up by runner and is in progress")
		return true, true, nil
	}

	for {
		now := workflow.Now(ctx)
		if now.After(overallTimeout) {
			l.Error("overall job timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout")
			tags["status"] = "overall_timeout"
			etags["status"] = "overall_timeout"
			w.mw.Event(ctx, &statsd.Event{
				Title:          "Overall job timeout reached without job starting",
				Text:           "Overall end-to-end job execution timeout reached without ever having been picked up",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-timeout-awaiting-job-pickup",
			})
			return false, false, nil
		}
		if now.After(availableTimeout) {
			l.Error("timeout waiting for job to be picked up")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "timeout waiting for runner to pick up job")
			tags["status"] = "available_timeout"
			etags["status"] = "available_timeout"
			w.mw.Event(ctx, &statsd.Event{
				Title:          "Timeout waiting for runner job to be picked up",
				Text:           "Job was marked as ready for execution, and runner appears to be in a healthy state, but runner did not pick up the job within the available timeout",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-timeout-awaiting-job-pickup",
			})
			return true, false, nil
		}

		// Wait for a state flag or until the earliest deadline fires so we
		// re-check the wall-clock timeouts at the top of the loop.
		deadline := availableTimeout
		if overallTimeout.Before(deadline) {
			deadline = overallTimeout
		}
		waitDuration := deadline.Sub(now)
		if waitDuration <= 0 {
			continue
		}
		if _, err := workflow.AwaitWithTimeout(ctx, waitDuration, state.anyFlag); err != nil {
			return false, false, err
		}

		if state.jobStatusChanged {
			state.jobStatusChanged = false
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
		if state.execCreated {
			state.execCreated = false
			execResp, err := activities.AwaitGetLatestJobExecution(ctx, activities.GetLatestJobExecutionRequest{
				JobID:       job.ID,
				AvailableAt: availableStart,
			})
			if err != nil {
				return false, false, fmt.Errorf("error fetching latest job execution: %w", err)
			}
			if execResp.Found {
				l.Info("job picked up by runner and is in progress")
				return true, true, nil
			}
		}
		// Clear the rest of the flags.
		state.execStatusChanged = false
		state.runnerStatusChanged = false
		state.runnerRestarted = false
	}
}
