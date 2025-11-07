package jobloop

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/sourcegraph/conc/panics"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/pkg/retry"
)

func (j *jobLoop) updateJobExecutionStatus(ctx context.Context, jobID, jobExecutionID string, status models.AppRunnerJobExecutionStatus) error {
	fn := func(ctx context.Context) error {
		if _, err := j.apiClient.UpdateJobExecution(ctx, jobID, jobExecutionID, &models.ServiceUpdateRunnerJobExecutionRequest{
			Status: status,
		}); err != nil {
			return fmt.Errorf("unable to update job execution status: %w", err)
		}

		return nil
	}

	if err := retry.Retry(ctx, fn, retry.WithMaxAttempts(10), retry.WithSleep(5)); err != nil {
		return err
	}

	return nil
}

func (j *jobLoop) errToStatus(err error) models.AppRunnerJobExecutionStatus {
	if errors.Is(err, context.DeadlineExceeded) {
		return models.AppRunnerJobExecutionStatusTimedDashOut
	}

	return models.AppRunnerJobExecutionStatusFailed
}

func (j *jobLoop) execJobStep(ctx context.Context, l *zap.Logger, logProvider *log.LoggerProvider, step *executeJobStep, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	l = l.With(zap.String("runner_job_execution_step.name", step.name))
	ctx = pkgctx.SetLogger(ctx, l)

	startTS := time.Now()
	tags := metrics.ToTags(map[string]string{})

	if err := j.updateJobExecutionStatus(ctx, job.ID, jobExecution.ID, step.startStatus); err != nil {
		j.mw.Incr("job_step", metrics.AddTagsMap(tags, map[string]string{
			"status":   "error",
			"err_type": "update_job_execution",
		}))
		j.mw.Timing("job_step.duration", time.Since(startTS), metrics.AddTagsMap(tags, map[string]string{
			"status":   "error",
			"err_type": "update_job_execution",
		}))
		return err
	}

	var (
		pc  panics.Catcher
		err error
	)
	pc.Try(func() {
		err = step.fn(ctx, step.handler, job, jobExecution)
	})

	// when a job handler panics, we update the job to a failed status, and propagate the error
	recovered := pc.Recovered()
	if recovered != nil {
		status := models.AppRunnerJobExecutionStatusFailed
		if updateErr := j.updateJobExecutionStatus(ctx, job.ID, jobExecution.ID, status); updateErr != nil {
			j.errRecorder.Record("update_job_execution", updateErr)
		}

		j.mw.Incr("job_step", metrics.AddTagsMap(tags, map[string]string{
			"status":   "error",
			"err_type": "panic",
		}))
		j.mw.Timing("job_step.duration", time.Since(startTS), metrics.AddTagsMap(tags, map[string]string{
			"status":   "error",
			"err_type": "panic",
		}))

		l.Error("panic in " + step.name)
		l.Error(recovered.String())
		l.Error(string(debug.Stack()))

		if flushErr := logProvider.ForceFlush(ctx); flushErr != nil {
			if !errors.Is(flushErr, context.Canceled) {
				l.Error("unable to flush logger during panic", zap.Error(flushErr))
			}
		}

		panic(recovered)
	}

	if err == nil {
		l.Info("step was completed successfully", zap.String("step", step.name))
		j.mw.Incr("job_step", metrics.AddTagsMap(tags, map[string]string{
			"status": "ok",
		}))
		j.mw.Timing("job_step.duration", time.Since(startTS), metrics.AddTagsMap(tags, map[string]string{
			"status": "ok",
		}))
		return nil
	}

	l.Error("job step errored "+err.Error(), zap.String("step", step.name), zap.Error(err))

	// handle the error by cleaning up the execution using the handler.
	status := j.errToStatus(err)
	if updateErr := j.updateJobExecutionStatus(ctx, job.ID, jobExecution.ID, status); updateErr != nil {
		j.errRecorder.Record("update_job_execution", updateErr)
	}

	if step.cleanupFn == nil {
		return err
	}
	if cleanupErr := step.cleanupFn(ctx, step.handler, job, jobExecution); cleanupErr != nil {
		j.errRecorder.Record("cleanup", cleanupErr)
	}

	j.mw.Incr("job_step", metrics.AddTagsMap(tags, map[string]string{
		"status":   "error",
		"err_type": "handler",
	}))
	j.mw.Timing("job_step.duration", time.Since(startTS), metrics.AddTagsMap(tags, map[string]string{
		"status":   "error",
		"err_type": "handler",
	}))
	return err
}
