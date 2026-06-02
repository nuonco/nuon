package job

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"

	processjobsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processjob"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
)

const (
	pollJobPeriod              time.Duration = time.Second * 10
	pollJobMaxWorkflowDuration               = time.Minute * 2
)

var failureStatuses = []app.RunnerJobStatus{
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

// @temporal-gen-v2 workflow
// @execution-timeout 1h
// @task-timeout 1m
// @task-queue "api"
// @id-generator WorkflowIDCallback
func (w *Workflows) ExecuteJob(ctx workflow.Context, req *ExecuteJobRequest) (app.RunnerJobStatus, error) {
	if workflow.GetInfo(ctx).ContinuedExecutionRunID != "" {
		return w.pollJob(ctx, req)
	}

	cb := callback.New(ctx, req.JobID)
	queueSignalID, err := w.queueJob(ctx, req.RunnerID, req.JobID, cb)
	if err != nil {
		return app.RunnerJobStatusUnknown, errors.Wrap(err, "unable to queue job")
	}

	if queueSignalID != "" {
		// Queue path: await signal completion via callback.
		if _, err := callback.Await(ctx, cb); err != nil {
			return app.RunnerJobStatusUnknown, errors.Wrap(err, "queue signal failed")
		}
	} else {
		if _, err := w.pollJob(ctx, req); err != nil {
			return app.RunnerJobStatus(""), err
		}
	}

	job, err := activities.AwaitPkgWorkflowsJobGetJobByID(ctx, req.JobID)
	if err != nil {
		return app.RunnerJobStatus(""), errors.Wrap(err, "unable to get job")
	}

	if generics.SliceContains(job.Status, failureStatuses) {
		msg := "job did not succeed"
		if job.StatusDescription != "" {
			msg = fmt.Sprintf("job did not succeed: %s", job.StatusDescription)
		}
		return job.Status, temporal.NewNonRetryableApplicationError(msg, "api", fmt.Errorf("job failed with status %s: %s", job.Status, job.StatusDescription))
	}

	return job.Status, nil
}

// queueJob dispatches the job for execution. Returns the queue signal ID when the
// queue path is used (non-empty string), or empty string for the poll path.
func (j *Workflows) queueJob(ctx workflow.Context, runnerID, jobID string, cb callback.Ref) (string, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return "", errors.Wrap(err, "expected a log stream in the context to poll job")
	}

	// Check if this runner uses queue-based job dispatch (parallel-runner-jobs feature flag).
	// If a per-job-group queue exists, enqueue the processjob signal directly.
	queueResp, err := activities.AwaitPkgWorkflowsJobGetRunnerJobGroupQueue(ctx, &activities.GetRunnerJobGroupQueueRequest{
		RunnerID: runnerID,
		JobID:    jobID,
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to check runner job group queue")
	}

	if queueResp.QueueID != "" {
		l.Info("queueing job via job-group queue", zap.String("runner-id", runnerID), zap.String("queue-id", queueResp.QueueID))
		enqueueResp, err := queueclient.AwaitEnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   queueResp.QueueID,
			Signal:    &processjobsignal.Signal{RunnerID: runnerID, JobID: jobID},
			OwnerID:   jobID,
			OwnerType: "runner_jobs",
			Callback:  cb,
		})
		if err != nil {
			return "", errors.Wrap(err, "unable to enqueue job to queue")
		}
		return enqueueResp.ID, nil
	}

	return "", errors.New("no job-group queue found for runner")
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

	var cancelled bool
	donechan := ctx.Done()
	dctx, cancel := workflow.NewDisconnectedContext(ctx)
	defer func() {
		if cancelled {
			cancel()
		}
	}()

	workflow.Go(dctx, func(ctx workflow.Context) {
		donechan.Receive(ctx, nil)
		cancelled = true
		err = activities.AwaitPkgWorkflowsJobCancelJobByID(ctx, jobID)
		if err != nil {
			l.Error("workflow context was cancelled, but propagating cancellation to job failed", zap.Error(err))
		}
	})

	for {
		// if the job is already timed out, there is no reason to continue.
		now := workflow.Now(dctx)
		if now.After(job.CreatedAt.Add(job.OverallTimeout)) {
			return app.RunnerJobStatusTimedOut, temporal.NewNonRetryableApplicationError("overall timeout reached", "api", fmt.Errorf("timeout"))
		}

		if now.After(startTS.Add(pollJobMaxWorkflowDuration)) {
			return job.Status, workflow.NewContinueAsNewError(dctx, workflow.GetInfo(dctx).WorkflowType.Name, req)
		}

		job, err := activities.AwaitPkgWorkflowsJobGetJobByID(dctx, jobID)
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
			msg := "job did not succeed"
			if job.StatusDescription != "" {
				msg = fmt.Sprintf("job did not succeed: %s", job.StatusDescription)
			}
			return job.Status, temporal.NewNonRetryableApplicationError(msg, "api", fmt.Errorf("job failed with status %s: %s", job.Status, job.StatusDescription))
		}

		if job.Status == app.RunnerJobStatusQueued {
			if err := j.logJobQueue(dctx, jobID); err != nil {
				return app.RunnerJobStatusUnknown, errors.Wrap(err, "unable to get runner job queue")
			}
		}

		donechan.ReceiveWithTimeout(ctx, pollJobPeriod, nil)
	}
}
