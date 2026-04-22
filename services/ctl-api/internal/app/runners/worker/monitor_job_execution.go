package worker

import (
	"errors"
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

// monitorJobExecution watches an in-flight job execution until it reaches a
// terminal state or a wall-clock timeout is hit.
//
// Prior to the signal-based rewrite, this function polled four activities
// every second (heartbeat, job status, runner status, execution status),
// generating ~15 history events per second. On long-running or stuck jobs
// that blew out workflow history and drove Temporal lock contention.
//
// The loop now blocks on a Selector: two wall-clock timers plus the
// processjob wake-up signal channel. Writers of the monitored state fire
// the signal when they change it, which wakes the workflow for one
// authoritative re-read. Happy-path jobs produce a handful of history
// events instead of thousands.
//
// Authoritative reads on wake-up are intentional — the signal payload is
// treated as a hint, never trusted blindly. That matches the old loop's
// semantics and keeps the workflow correct under out-of-order signal
// delivery.
func (w *Workflows) monitorJobExecution(ctx workflow.Context, job *app.RunnerJob) (bool, error) {
	startTS := workflow.Now(ctx)
	tags := map[string]string{
		"status":    "ok",
		"job_type":  string(job.Type),
		"job_group": string(job.Group),
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
		w.mw.Incr(ctx, "runner.job_execution", metrics.ToTags(tags)...)
		e2eLatency := workflow.Now(ctx).Sub(startTS)
		w.mw.Timing(ctx, "runner.job_execution.latency", e2eLatency, metrics.ToTags(tags)...)
	}()

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return false, err
	}

	jobExecution, err := activities.AwaitGetCurrentJobExecutionByJobID(ctx, job.ID)
	if err != nil {
		return false, fmt.Errorf("error fetching latest job execution: %w", err)
	}

	executionTimeout := jobExecution.CreatedAt.Add(job.ExecutionTimeout)
	overallTimeoutAt := job.CreatedAt.Add(job.OverallTimeout)

	signalChan := workflow.GetSignalChannel(ctx, processjobsignals.SignalName)

	for {
		retry, done, err := w.checkMonitorState(ctx, job, jobExecution, tags, etags, l)
		if err != nil || done {
			return retry, err
		}

		// Wait for either a wake-up signal or one of the wall-clock deadlines.
		// Any signal wakes us for a re-check above; timers short-circuit to
		// the timeout branches.
		sel := workflow.NewSelector(ctx)

		var wakeup processjobsignals.WakeUp
		sel.AddReceive(signalChan, func(c workflow.ReceiveChannel, _ bool) {
			c.Receive(ctx, &wakeup)
		})

		now := workflow.Now(ctx)
		overallTimedOut := false
		executionTimedOut := false

		if d := overallTimeoutAt.Sub(now); d > 0 {
			sel.AddFuture(workflow.NewTimer(ctx, d), func(workflow.Future) {
				overallTimedOut = true
			})
		} else {
			overallTimedOut = true
		}

		if d := executionTimeout.Sub(now); d > 0 {
			sel.AddFuture(workflow.NewTimer(ctx, d), func(workflow.Future) {
				executionTimedOut = true
			})
		} else {
			executionTimedOut = true
		}

		sel.Select(ctx)

		if overallTimedOut {
			l.Error("overall timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "overall timeout")
			w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusTimedOut)
			tags["status"] = "overall_timeout"

			etags["job_status"] = string(app.RunnerJobStatusTimedOut)
			maps.Copy(etags, tags)
			w.mw.Event(ctx, &statsd.Event{
				Title:          "Overall timeout reached while job executing",
				Text:           "Overall end-to-end job execution timeout reached while waiting for job to bewcome healthy",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-timeout-while-executing",
			})
			return false, nil
		}

		if executionTimedOut {
			l.Error("execution timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "execution timeout")
			w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusTimedOut)
			tags["status"] = "execution_timeout"

			etags["job_status"] = string(app.RunnerJobStatusTimedOut)
			maps.Copy(etags, tags)
			w.mw.Event(ctx, &statsd.Event{
				Title:          "Overall timeout reached while job executing",
				Text:           "Overall end-to-end job execution timeout reached while waiting for job to become healthy",
				Tags:           metrics.ToTags(etags),
				SourceTypeName: "nuon-jobsys",
				Priority:       statsd.Normal,
				AlertType:      statsd.Error,
				AggregationKey: "runner-job-timeout-while-executing",
			})
			return true, nil
		}

		l.Debug("monitor wakeup", zap.String("reason", wakeup.Reason))
	}
}

// checkMonitorState runs one authoritative read of the monitored state and
// returns (retry, done, err). When done is true, the outer loop exits with
// the returned retry value. When done is false, the outer loop enters the
// selector and waits for the next wake-up.
func (w *Workflows) checkMonitorState(
	ctx workflow.Context,
	job *app.RunnerJob,
	jobExecution *app.RunnerJobExecution,
	tags, etags map[string]string,
	l *zap.Logger,
) (bool, bool, error) {
	hb, err := activities.AwaitGetMostRecentHeartBeatRequestByRunnerID(ctx, job.RunnerID)
	if err != nil {
		return false, true, err
	}
	if hb == nil {
		return false, true, errors.New("no heart beats found")
	}

	maxAliveTime := jobExecution.CreatedAt.Add(time.Minute)
	if hb.StartedAt.After(maxAliveTime) {
		l.Error(
			"runner restarted while job was in flight. job will be cancelled.",
			zap.Time("runner.started_at", hb.StartedAt),
			zap.Time("job_execution.created_at", jobExecution.CreatedAt),
		)
		w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusCancelled)

		maps.Copy(etags, tags)
		w.mw.Event(ctx, &statsd.Event{
			Title:          "Runner restarted while job in flight",
			Text:           "A runner was marked unhealthy during the job execution. The job will NOT be resumed if/when the runner recovers",
			Tags:           metrics.ToTags(etags),
			SourceTypeName: "nuon-jobsys",
			Priority:       statsd.Normal,
			AlertType:      statsd.Error,
			AggregationKey: "runner-job-dropped",
		})
	}

	jobStatus, err := activities.AwaitGetJobStatusByID(ctx, job.ID)
	if err != nil {
		return false, true, err
	}
	if jobStatus == app.RunnerJobStatusCancelled {
		l.Error("job was cancelled")
		w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusCancelled)
		tags["status"] = "cancelled"
		return true, true, nil
	}

	runnerStatus, err := activities.AwaitGetRunnerStatusByID(ctx, job.RunnerID)
	if err != nil {
		return false, true, err
	}
	if runnerStatus != app.RunnerStatusActive {
		l.Error("runner marked unhealthy during job")
		w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "runner became unhealthy during job")
		w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusFailed)
		tags["status"] = "runner_unhealthy"

		maps.Copy(etags, tags)
		w.mw.Event(ctx, &statsd.Event{
			Title:          "Runner marked unhealthy during job",
			Text:           "A runner was marked unhealthy during the job execution. The job will NOT be resumed if/when the runner recovers",
			Tags:           metrics.ToTags(etags),
			SourceTypeName: "nuon-jobsys",
			Priority:       statsd.Normal,
			AlertType:      statsd.Error,
			AggregationKey: "runner-job-dropped",
		})
		return true, true, nil
	}

	executionStatus, err := activities.AwaitGetJobExecutionStatus(ctx, activities.GetJobExecutionStatusRequest{
		JobExecutionID: jobExecution.ID,
	})
	if err != nil {
		return false, true, err
	}

	switch executionStatus {
	case app.RunnerJobExecutionStatusFinished:
		l.Info("job execution successfully finished")
		w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFinished, "finished")
		tags["status"] = "ok"
		return false, true, nil
	case app.RunnerJobExecutionStatusCancelled:
		l.Info("job cancelled")
		tags["status"] = "execution_cancelled"
		return true, true, nil
	case app.RunnerJobExecutionStatusFailed:
		l.Info("job execution failed")
		w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "failed")
		tags["status"] = "execution_failed"
		return true, true, nil
	case app.RunnerJobExecutionStatusTimedOut:
		l.Info("job execution timed out")
		w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "execution timed out")
		tags["status"] = "execution_timed_out"
		return true, true, nil
	case app.RunnerJobExecutionStatusNotAttempted:
		l.Info("job execution not attempted")
		w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "execution not attempted")
		tags["status"] = "execution_not_attempted"
		return true, true, nil
	}

	return false, false, nil
}
