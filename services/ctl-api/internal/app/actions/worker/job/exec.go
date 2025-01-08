package job

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

const (
	pollJobPeriod time.Duration = time.Second * 10
)

var failureStatuses = []app.RunnerJobStatus{
	app.RunnerJobStatusFailed,
	app.RunnerJobStatusTimedOut,
	app.RunnerJobStatusCancelled,
	app.RunnerJobStatusNotAttempted,
	app.RunnerJobStatusUnknown,
}

type ExecuteJobRequest struct {
	JobID      string `json:"job_id" validate:"required"`
	WorkflowID string `json:"workflow_id" validate:"required"`
}

// @temporal-gen workflow
// @execution-timeout 1h
// @task-timeout 1m
// @task-queue "api"
// @id-callback WorkflowIDCallback
func ExecuteJob(ctx workflow.Context, req *ExecuteJobRequest) (app.RunnerJobStatus, error) {
	jw := &jobWorkflow{}

	return jw.pollJob(ctx, req.JobID)
}

func (j *jobWorkflow) pollJob(ctx workflow.Context, jobID string) (app.RunnerJobStatus, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return app.RunnerJobStatusUnknown, errors.Wrap(err, "expected a log stream in the context to poll job")
	}

	job, err := activities.AwaitGetJobByID(ctx, jobID)
	if err != nil {
		return app.RunnerJobStatusUnknown, errors.Wrap(err, "unable to get job and set timeout")
	}

	for {
		// if the job is already timed out, there is no reason to continue. In some reasons, if a job fails and
		// is not updated by it's event loop, then this would catch that.
		now := workflow.Now(ctx)
		if now.After(job.CreatedAt.Add(job.OverallTimeout)) {
			return app.RunnerJobStatusTimedOut, temporal.NewNonRetryableApplicationError("overall timeout reached", "api", fmt.Errorf("timeout"))
		}

		job, err := activities.AwaitGetJobByID(ctx, jobID)
		if err != nil {
			return app.RunnerJobStatusUnknown, fmt.Errorf("unable to get job from database: %w", err)
		}

		if job.Status == app.RunnerJobStatusFinished {
			l.Info("job completed successfully")
			return job.Status, nil
		}

		// handle failure states here
		if generics.SliceContains(job.Status, failureStatuses) {
			l.Error(fmt.Sprintf("job failed with %s status", job.Status), zap.Any("status", job.Status))
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
