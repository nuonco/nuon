package worker

import (
	"errors"
	"fmt"
	"maps"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

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

	// poll the job execution, until it's completed
	executionTimeout := jobExecution.CreatedAt.Add(job.ExecutionTimeout)
	for {
		workflow.Sleep(ctx, defaultJobPollPeriod)

		now := workflow.Now(ctx)

		// when the overall timeout is hit, we mark both the runner job and execution as timed out
		// this is not retryable.
		if now.After(job.CreatedAt.Add(job.OverallTimeout)) {
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

		// when the execution timeout is hit, we mark both the runner job and execution as timed out
		// this is retryable
		if now.After(executionTimeout) {
			l.Error("execution timeout reached")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusTimedOut, "execution timeout")
			w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusTimedOut)
			tags["status"] = "execution_timeout"

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
			return true, nil
		}

		// if the runner was started after this execution was created, we mark the execution as in error
		// this is retryable
		hb, err := activities.AwaitGetMostRecentHeartBeatRequestByRunnerID(ctx, job.RunnerID)
		if err != nil {
			return false, err
		}
		if hb == nil {
			return false, errors.New("no heart beats found")
		}

		if hb.StartedAt.After(jobExecution.CreatedAt) {
			l.Error(
				"runner restarted while job was in flight. job will be cancelled.",
				zap.Time("runner.started_at", hb.StartedAt),
				zap.Time("job_execution.created_at", jobExecution.CreatedAt),
			)
			w.updateStatus(ctx, job.RunnerID, app.RunnerStatusError, "runner restarted while job was in flight")
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
			return false, err
		}
		if jobStatus == app.RunnerJobStatusCancelled {
			l.Error("job was cancelled")
			w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusCancelled)
			tags["status"] = "cancelled"
			return true, nil
		}

		// if the runner is deemed unhealthy, the job execution is marked as unknown, and the job is marked as
		// not attempted with the correct status, this is retryable.
		runnerStatus, err := activities.AwaitGetRunnerStatusByID(ctx, job.RunnerID)
		if runnerStatus != app.RunnerStatusActive {
			l.Error("runner marked unhealthy during job")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusUnknown, "runner became unhealthy during job")
			w.updateJobExecutionStatus(ctx, jobExecution.ID, app.RunnerJobExecutionStatusUnknown)
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
			return true, nil
		}

		// check the runner to make sure it did not become unhealthy from the time it picked up the execution,
		// to the current time
		executionStatus, err := activities.AwaitGetJobExecutionStatus(ctx, activities.GetJobExecutionStatusRequest{
			JobExecutionID: jobExecution.ID,
		})
		if err != nil {
			return false, err
		}

		// handle the job execution status
		switch executionStatus {
		case app.RunnerJobExecutionStatusFinished:
			l.Info("job execution successfully finished")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFinished, "finished")
			tags["status"] = "ok"
			return false, nil
		case app.RunnerJobExecutionStatusCancelled:
			l.Info("job cancelled")
			tags["status"] = "execution_cancelled"
			return true, nil
		case app.RunnerJobExecutionStatusFailed:
			l.Info("job execution failed")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "failed")
			tags["status"] = "execution_failed"
			return true, nil
		case app.RunnerJobExecutionStatusTimedOut:
			l.Info("job execution timed out")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "execution timed out")
			tags["status"] = "execution_timed_out"
			return true, nil
		case app.RunnerJobExecutionStatusNotAttempted:
			l.Info("job execution not attempted")
			w.updateJobStatus(ctx, job.ID, app.RunnerJobStatusFailed, "execution not attempted")
			tags["status"] = "execution_not_attempted"
			return true, nil
		default:
			continue
		}
	}
}
