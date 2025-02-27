package job

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"

	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job/activities"
)

const (
	pollJobPeriod              time.Duration = time.Second * 10
	pollJobMaxWorkflowDuration               = time.Minute * 2
)

var failureStatuses = []app.RunnerJobStatus{
	app.RunnerJobStatusFailed,
	app.RunnerJobStatusFailed,
	app.RunnerJobStatusTimedOut,
	app.RunnerJobStatusCancelled,
	app.RunnerJobStatusNotAttempted,
	app.RunnerJobStatusUnknown,
}

type ExecuteJobRequest struct {
	RunnerID   string `json:"runner_id" validate:"required"`
	JobID      string `json:"job_id" validate:"required"`
	WorkflowID string `json:"workflow_id" validate:"required"`
}

func AwaitExecuteJob(ctx workflow.Context, req *ExecuteJobRequest) (app.RunnerJobStatus, error) {
	var w Workflows
	return (&w).AwaitExecuteJob(ctx, req)
}

// @temporal-gen workflow
// @execution-timeout 1h
// @task-timeout 1m
// @task-queue "api"
// @id-callback WorkflowIDCallback
func (w *Workflows) ExecuteJob(ctx workflow.Context, req *ExecuteJobRequest) (app.RunnerJobStatus, error) {
	if workflow.GetInfo(ctx).ContinuedExecutionRunID != "" {
		return w.pollJob(ctx, req)
	}

	if err := w.queueJob(ctx, req.RunnerID, req.JobID); err != nil {
		return app.RunnerJobStatusUnknown, errors.Wrap(err, "unable to queue job")
	}

	if _, err := w.pollJob(ctx, req); err != nil {
		return app.RunnerJobStatus(""), err
	}

	return app.RunnerJobStatusUnknown, nil
}

func (j *Workflows) queueJob(ctx workflow.Context, runnerID, jobID string) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "expected a log stream in the context to poll job")
	}

	// now queue and execute the job
	l.Info("queueing job on runner event loop", zap.String("runner-id", runnerID))

	j.evClient.Send(ctx, runnerID, &runnersignals.Signal{
		Type:  runnersignals.OperationProcessJob,
		JobID: jobID,
	})

	return nil
}

func (j *Workflows) pollJob(ctx workflow.Context, req *ExecuteJobRequest) (app.RunnerJobStatus, error) {
	jobID := req.JobID
	wkflowInfo := workflow.GetInfo(ctx)
	job, err := activities.AwaitPkgWorkflowsJobGetJobByID(ctx, jobID)
	if err != nil {
		return app.RunnerJobStatusUnknown, errors.Wrap(err, "unable to get job and set timeout")
	}

	defaultTags := map[string]string{
		"namespace": wkflowInfo.Namespace,
		"status":    "ok",
		"job_type":  string(job.Type),
	}
	startTS := workflow.Now(ctx)
	defer func() {
		j.mw.Incr(ctx, "runner_job.client.incr", metrics.ToTags(defaultTags)...)
		e2eLatency := workflow.Now(ctx).Sub(startTS)
		j.mw.Timing(ctx, "runner_job.client.latency", e2eLatency, metrics.ToTags(defaultTags)...)
	}()

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return app.RunnerJobStatusUnknown, errors.Wrap(err, "expected a log stream in the context to poll job")
	}

	// fetch starting queued
	queued, err := activities.AwaitPkgWorkflowsJobGetRunnerJobQueueByJobID(ctx, jobID)
	if err != nil {
		return "", errors.Wrap(err, "unable to get runner job queue")
	}
	j.mw.Gauge(ctx, "runner_job.client.starting_queue", float64(len(queued)), metrics.ToTags(defaultTags)...)

	for {
		// if the job is already timed out, there is no reason to continue. In some reasons, if a job fails and
		// is not updated by it's event loop, then this would catch that.
		now := workflow.Now(ctx)
		if now.After(job.CreatedAt.Add(job.OverallTimeout)) {
			return app.RunnerJobStatusTimedOut, temporal.NewNonRetryableApplicationError("overall timeout reached", "api", fmt.Errorf("timeout"))
		}

		if now.After(startTS.Add(pollJobMaxWorkflowDuration)) {
			return job.Status, workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, req)
		}

		job, err := activities.AwaitPkgWorkflowsJobGetJobByID(ctx, jobID)
		if err != nil {
			return app.RunnerJobStatusUnknown, fmt.Errorf("unable to get job from database: %w", err)
		}

		if job.Status == app.RunnerJobStatusFinished {
			l.Info("job completed successfully")
			return job.Status, nil
		}

		// handle failure states here
		if generics.SliceContains(job.Status, failureStatuses) {
			l.Error(fmt.Sprintf("job failed with status (%s) (%s)", job.Status, job.StatusDescription),
				zap.Any("status", job.Status),
				zap.Any("status_description", job.StatusDescription),
			)
			return job.Status, temporal.NewNonRetryableApplicationError("job did not succeed", "api", fmt.Errorf("job failed with status %s", job.Status))
		}

		if job.Status == app.RunnerJobStatusQueued {
			if err := j.logJobQueue(ctx, jobID); err != nil {
				return app.RunnerJobStatusUnknown, errors.Wrap(err, "unable to get runner job queue")
			}
		}

		workflow.Sleep(ctx, pollJobPeriod)
	}
}
