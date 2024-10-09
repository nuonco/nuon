package jobloop

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/slog"
)

type executeJobStep struct {
	name      string
	fn        func(context.Context, jobs.JobHandler, *models.AppRunnerJob, *models.AppRunnerJobExecution) error
	cleanupFn func(context.Context, jobs.JobHandler, *models.AppRunnerJob, *models.AppRunnerJobExecution) error
	handler   jobs.JobHandler

	startStatus models.AppRunnerJobExecutionStatus
}

func (j *jobLoop) executeJob(ctx context.Context, job *models.AppRunnerJob) error {
	l := slog.DefaultLogger(j.loggerProvider, j.settings, j.cfg, slog.LoggerTypeJob)
	l = l.With("runner_job.id", job.ID)

	// create an execution in the API
	l.Info("creating job execution")
	execution, err := j.apiClient.CreateJobExecution(ctx, job.ID, new(models.ServiceCreateRunnerJobExecutionRequest))
	if err != nil {
		return errors.Wrap(err, "unable to create execution")
	}
	l = l.With("runner_job_execution.id", execution.ID)

	l.Info("getting job handler")
	handler, err := j.getHandler(job)
	if err != nil {
		if err := j.updateJobExecutionStatus(ctx, job.ID, execution.ID, models.AppRunnerJobExecutionStatusFailed); err != nil {
			j.errRecorder.Record("no handler found", err)
		}

		return err
	}

	steps, err := j.getJobSteps(ctx, handler)
	if err != nil {
		return errors.Wrap(err, "unable to get job steps")
	}

	for _, step := range steps {
		l.Info("executing job step", "step", step.name)
		if err := j.execJobStep(ctx, l, step, job, execution); err != nil {
			return errs.WithHandlerError(err, j.jobGroup, step.name, job.Type)
		}
	}

	if err := j.updateJobExecutionStatus(ctx, job.ID, execution.ID, models.AppRunnerJobExecutionStatusFinished); err != nil {
		return errors.Wrap(err, "unable to update job execution status after successful execution")
	}

	l.Info("finished job", "name", handler.Name())
	if err := j.loggerProvider.ForceFlush(ctx); err != nil {
		return errors.Wrap(err, "unable to flush logger")
	}

	return nil
}
