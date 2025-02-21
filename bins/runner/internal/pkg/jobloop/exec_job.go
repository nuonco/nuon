package jobloop

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-runner-go/models"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/errs"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
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
	jl, err := slog.NewOTELProvider(j.cfg, j.settings, job.LogStreamID)
	if err != nil {
		return errors.Wrap(err, "unable to create otel provider")
	}

	l, err := log.NewOTELJobLogger(j.cfg, jl)
	if err != nil {
		return errors.Wrap(err, "unable to get job logger")
	}

	l = l.With(zap.String("runner_job.id", job.ID))
	l = l.With(zap.String("log_stream.id", job.LogStreamID))

	// create an execution in the API
	l.Info("creating job execution")
	execution, err := j.apiClient.CreateJobExecution(ctx, job.ID, new(models.ServiceCreateRunnerJobExecutionRequest))
	if err != nil {
		return errors.Wrap(err, "unable to create execution")
	}
	l = l.With(zap.String("runner_job_execution.id", execution.ID))

	l.Info("getting job handler")
	handler, err := j.getHandler(job)
	if err != nil {
		l.Error("no valid job handler found for job",
			zap.String("type", string(job.Type)),
			zap.Error(err),
		)
		if err := j.updateJobExecutionStatus(ctx, job.ID, execution.ID, models.AppRunnerJobExecutionStatusFailed); err != nil {
			j.errRecorder.Record("no handler found", err)
		}

		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	doneCh := make(chan struct{})
	go func() {
		j.monitorJob(ctx, cancel, doneCh, job.ID, l)
	}()

	steps, err := j.getJobSteps(ctx, handler)
	if err != nil {
		return errors.Wrap(err, "unable to get job steps")
	}

	for _, step := range steps {
		l.Info("executing job step "+step.name, zap.String("step", step.name))
		if err := j.execJobStep(ctx, l, step, job, execution); err != nil {
			return errs.WithHandlerError(err, j.jobGroup, step.name, job.Type)
		}
	}

	if err := j.updateJobExecutionStatus(ctx, job.ID, execution.ID, models.AppRunnerJobExecutionStatusFinished); err != nil {
		return errors.Wrap(err, "unable to update job execution status after successful execution")
	}

	l.Info("finished job", zap.String("name", handler.Name()))
	if err := jl.ForceFlush(ctx); err != nil {
		return errors.Wrap(err, "unable to flush logger")
	}
	close(doneCh)

	return nil
}
