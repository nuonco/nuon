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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

// monitorJobExecution watches a running job. Replaces per-second polling
// with workflow.Await on the push-update flags in processJobState; wakes
// when an external state writer fires a relevant workflow update or when a
// wall-clock deadline fires.
func (w *Workflows) monitorJobExecution(ctx workflow.Context, state *processJobState, job *app.RunnerJob) (bool, error) {
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
	overallTimeout := job.CreatedAt.Add(job.OverallTimeout)

	// Initial-state read: an external update may have fired between the
	// exec row being created and our handler being registered. Check
	// terminal states up front.
	if retry, done, err := w.checkMonitorState(ctx, l, job, jobExecution, tags, etags); done || err != nil {
		return retry, err
	}

	for {
		now := workflow.Now(ctx)

		if now.After(overallTimeout) {
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

		if now.After(executionTimeout) {
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

		// Wait for an update flag or until the earliest wall-clock deadline
		// fires so we re-check the timeout branches at the top of the loop
		// even if no external update ever arrives.
		deadline := executionTimeout
		if overallTimeout.Before(deadline) {
			deadline = overallTimeout
		}
		waitDuration := deadline.Sub(now)
		if waitDuration <= 0 {
			continue
		}
		if _, err := workflow.AwaitWithTimeout(ctx, waitDuration, state.anyFlag); err != nil {
			return false, err
		}

		retry, done, err := w.handleMonitorFlags(ctx, l, state, job, jobExecution, tags, etags)
		if err != nil {
			return false, err
		}
		if done {
			return retry, nil
		}
	}
}

// checkMonitorState runs the full monitor decision logic once, re-reading
// authoritative state from the DB. Returns (retry, done, err) where done is
// true when a terminal state was reached.
func (w *Workflows) checkMonitorState(
	ctx workflow.Context,
	l *zap.Logger,
	job *app.RunnerJob,
	jobExecution *app.RunnerJobExecution,
	tags, etags map[string]string,
) (bool, bool, error) {
	// Heartbeat / restart check.
	hb, err := activities.AwaitGetMostRecentHeartBeatRequestByRunnerID(ctx, job.RunnerID)
	if err != nil {
		return false, false, err
	}
	if hb == nil {
		return false, false, errors.New("no heart beats found")
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
		return false, false, err
	}
	if jobStatus == app.RunnerJobStatusCancelled {
		l.Error("job was cancelled")
		w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusCancelled)
		tags["status"] = "cancelled"
		return true, true, nil
	}

	runnerStatus, err := activities.AwaitGetRunnerStatusByID(ctx, job.RunnerID)
	if err != nil {
		return false, false, err
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
		return false, false, err
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
	default:
		return false, false, nil
	}
}

// handleMonitorFlags runs when at least one update flag is set. It clears
// all flags, re-reads authoritative state, and returns whether to exit.
// The flag payloads are advisory only — every wake does a full re-read
// because state could have moved past what the triggering update reported.
func (w *Workflows) handleMonitorFlags(
	ctx workflow.Context,
	l *zap.Logger,
	state *processJobState,
	job *app.RunnerJob,
	jobExecution *app.RunnerJobExecution,
	tags, etags map[string]string,
) (bool, bool, error) {
	state.clear()
	return w.checkMonitorState(ctx, l, job, jobExecution, tags, etags)
}
