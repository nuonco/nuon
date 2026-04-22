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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/processjobsignals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

// startJobExecution marks a job as available and waits until:
//   - the runner becomes Active (non-operations jobs only), then
//   - the runner creates a RunnerJobExecution row for this job.
//
// Previously this used two nested 1-second poll loops, each calling
// AwaitGetRunnerStatusByID, AwaitGetJobStatusByID, and
// AwaitGetLatestJobExecution on every tick. Those loops generated ~15
// history events/s.
//
// Now both loops block on the processjob wake-up signal channel. Writers
// that mutate runner status, job status, or create the execution row send
// that signal, waking the workflow for a single authoritative re-read.
// The wall-clock deadlines are enforced by timer futures in the selector
// rather than by comparing workflow.Now() inside a busy loop.
func (w *Workflows) startJobExecution(ctx workflow.Context, job *app.RunnerJob) (bool, bool, error) {
	startTS := workflow.Now(ctx)
	tags := map[string]string{
		"status":        "ok",
		"job_type":      string(job.Type),
		"job_group":     string(job.Group),
		"job_operation": string(job.Operation),
	}

	etags := maps.Clone(tags)
	etags["job_id"] = job.ID
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

	overallDeadline := job.CreatedAt.Add(job.OverallTimeout)
	availableDeadline := availableStart.Add(job.AvailableTimeout)

	signalChan := workflow.GetSignalChannel(ctx, processjobsignals.SignalName)

	// Phase 1 (non-operations jobs only): wait for runner to be Active.
	if job.Group != app.RunnerJobGroupOperations {
		for {
			runnerStatus, err := activities.AwaitGetRunnerStatusByID(ctx, job.RunnerID)
			if err != nil {
				l.Warn("unable to determine runner status", zap.Error(err))
				return false, false, err
			}
			if runnerStatus == app.RunnerStatusActive {
				break
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

			etags["runner_status"] = string(runnerStatus)

			retry, done := w.waitForWakeup(ctx, signalChan, overallDeadline, availableDeadline, func() {
				l.Error("overall job timeout reached")
				w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout waiting for runner to be healthy")
				tags["status"] = "runner_unhealthy"
				w.mw.Event(ctx, &statsd.Event{
					Title:          "Overall job timeout reached waiting for runner to become healthy",
					Text:           "Overall end-to-end job execution timeout reached while waiting for job to bewcome healthy",
					Tags:           metrics.ToTags(etags),
					SourceTypeName: "nuon-jobsys",
					Priority:       statsd.Normal,
					AlertType:      statsd.Error,
					AggregationKey: "runner-job-timeout-waiting-for-healthy-runner",
				})
			}, func() {
				l.Error("timeout waiting for job to be picked up")
				w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "timeout waiting for runner to become healthy")
				tags["status"] = "runner_unhealthy"
				w.mw.Event(ctx, &statsd.Event{
					Title:          "Available timeout reached waiting for runner to become healthy",
					Text:           "Job is ready for execution, but runner did not become healthy within the available timeout",
					Tags:           metrics.ToTags(etags),
					SourceTypeName: "nuon-jobsys",
					Priority:       statsd.Low,
					AlertType:      statsd.Warning,
					AggregationKey: "runner-job-timeout-waiting-for-healthy-runner",
				})
			})
			if done {
				return retry, false, nil
			}
		}
	}

	// Phase 2: wait for the runner to create an execution row (job "picked up").
	for {
		jobStatus, err := activities.AwaitGetJobStatusByID(ctx, job.ID)
		if err != nil {
			return false, false, nil
		}
		if jobStatus == app.RunnerJobStatusCancelled {
			l.Error("job was cancelled")
			tags["status"] = "job_cancelled"
			return false, false, nil
		}

		execResp, err := activities.AwaitGetLatestJobExecution(ctx, activities.GetLatestJobExecutionRequest{
			JobID:       job.ID,
			AvailableAt: availableStart,
		})
		if err != nil {
			return false, false, fmt.Errorf("error fetching latest job execution: %w", err)
		}
		if execResp.Found {
			break
		}

		retry, done := w.waitForWakeup(ctx, signalChan, overallDeadline, availableDeadline, func() {
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
		}, func() {
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
		})
		if done {
			return retry, false, nil
		}
	}

	l.Info("job picked up by runner and is in progress")
	return true, true, nil
}

// waitForWakeup blocks on the signal channel or timeout futures.
// Returns (retry, done): when done=true the caller should return immediately.
// onOverall and onAvailable are called (for side-effects like status updates
// and metrics) before returning done=true on the respective timeout.
func (w *Workflows) waitForWakeup(
	ctx workflow.Context,
	signalChan workflow.ReceiveChannel,
	overallDeadline, availableDeadline time.Time,
	onOverall, onAvailable func(),
) (retry bool, done bool) {
	sel := workflow.NewSelector(ctx)

	var wakeup processjobsignals.WakeUp
	sel.AddReceive(signalChan, func(c workflow.ReceiveChannel, _ bool) {
		c.Receive(ctx, &wakeup)
	})

	now := workflow.Now(ctx) // time.Time from Temporal
	overallTimedOut := false
	availableTimedOut := false

	if d := overallDeadline.Sub(now); d > 0 {
		sel.AddFuture(workflow.NewTimer(ctx, d), func(workflow.Future) {
			overallTimedOut = true
		})
	} else {
		overallTimedOut = true
	}

	if d := availableDeadline.Sub(now); d > 0 {
		sel.AddFuture(workflow.NewTimer(ctx, d), func(workflow.Future) {
			availableTimedOut = true
		})
	} else {
		availableTimedOut = true
	}

	sel.Select(ctx)

	if overallTimedOut {
		onOverall()
		return false, true
	}
	if availableTimedOut {
		onAvailable()
		return true, true
	}
	return false, false
}
