package jobloop

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"github.com/sourcegraph/conc/panics"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/retry"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

// jobExecutionStatusDescriptionMaxLen caps the error description sent to the API
// so a long stack trace doesn't bloat the stored status history.
const jobExecutionStatusDescriptionMaxLen = 2048

const (
	jobExecutionResultWriteTimeout         = 30 * time.Second
	jobExecutionTerminalStatusWriteTimeout = 30 * time.Second
)

// writeJobExecutionStatus is the synchronous, retry-wrapped API call.
// It's the writer the coalescer's background goroutine drives and also
// the call terminal updates fall through to directly. Intermediate
// (non-terminal) callers go through the coalescer instead — see
// `statusCoalescer` in status_coalescer.go.
func (j *jobLoop) writeJobExecutionStatus(ctx context.Context, jobID, jobExecutionID string, status models.AppRunnerJobExecutionStatus, description string) error {
	if len(description) > jobExecutionStatusDescriptionMaxLen {
		description = description[:jobExecutionStatusDescriptionMaxLen] + "…(truncated)"
	}
	fn := func(ctx context.Context) error {
		if _, err := j.apiClient.UpdateJobExecution(ctx, jobID, jobExecutionID, &models.ServiceUpdateRunnerJobExecutionRequest{
			Status:            status,
			StatusDescription: description,
		}); err != nil {
			return fmt.Errorf("unable to update job execution status: %w", err)
		}

		return nil
	}

	if err := retry.Retry(ctx, fn, retry.WithMaxAttempts(10), retry.WithSleep(5*time.Second)); err != nil {
		return err
	}

	return nil
}

// updateJobExecutionStatus and updateJobExecutionStatusWithDescription
// are the legacy synchronous entry points. They route through the
// per-execution coalescer when one is attached so the runner doesn't
// block on intermediate transition pings, and fall back to a direct
// synchronous write when there isn't one (e.g. early failure before
// `executeJob` started the coalescer).
func (j *jobLoop) updateJobExecutionStatus(ctx context.Context, jobID, jobExecutionID string, status models.AppRunnerJobExecutionStatus) error {
	return j.updateJobExecutionStatusWithDescription(ctx, jobID, jobExecutionID, status, "")
}

func (j *jobLoop) updateJobExecutionStatusWithDescription(ctx context.Context, jobID, jobExecutionID string, status models.AppRunnerJobExecutionStatus, description string) error {
	if isTerminalExecutionStatus(status) {
		statusCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), jobExecutionTerminalStatusWriteTimeout)
		defer cancel()

		if c := j.coalescerFor(jobExecutionID); c != nil {
			return c.WriteTerminal(statusCtx, status, description)
		}
		return j.writeJobExecutionStatus(statusCtx, jobID, jobExecutionID, status, description)
	}

	if c := j.coalescerFor(jobExecutionID); c != nil {
		c.EnqueueNonTerminal(status, description)
		return nil
	}
	return j.writeJobExecutionStatus(ctx, jobID, jobExecutionID, status, description)
}

// coalescerFor returns the coalescer attached to the current execution,
// or nil if none has been registered yet. The map is keyed by execution
// id so concurrent jobs (parallel-runner-jobs feature) don't share a
// writer.
func (j *jobLoop) coalescerFor(executionID string) *statusCoalescer {
	j.coalescersMu.Lock()
	defer j.coalescersMu.Unlock()
	return j.coalescers[executionID]
}

func (j *jobLoop) attachCoalescer(executionID string, c *statusCoalescer) {
	j.coalescersMu.Lock()
	defer j.coalescersMu.Unlock()
	if j.coalescers == nil {
		j.coalescers = make(map[string]*statusCoalescer)
	}
	j.coalescers[executionID] = c
}

func (j *jobLoop) detachCoalescer(executionID string) {
	j.coalescersMu.Lock()
	defer j.coalescersMu.Unlock()
	delete(j.coalescers, executionID)
}

func (j *jobLoop) errToStatus(err error) models.AppRunnerJobExecutionStatus {
	if errors.Is(err, context.DeadlineExceeded) {
		return models.AppRunnerJobExecutionStatusTimedDashOut
	}

	return models.AppRunnerJobExecutionStatusFailed
}

func (j *jobLoop) writeFallbackJobExecutionResult(ctx context.Context, job *models.AppRunnerJob, execution *models.AppRunnerJobExecution, handler, step string, jobErr error) error {
	resultCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), jobExecutionResultWriteTimeout)
	defer cancel()

	req := &models.ServiceCreateRunnerJobExecutionResultRequest{
		Success:   false,
		ErrorCode: 0,
		ErrorMetadata: map[string]string{
			"step":     step,
			"handler":  handler,
			"job_type": string(job.Type),
			"message":  jobErr.Error(),
		},
	}
	writeResult := func(ctx context.Context) error {
		if _, err := j.apiClient.CreateJobExecutionResult(ctx, job.ID, execution.ID, req); err != nil {
			return fmt.Errorf("unable to create fallback job execution result: %w", err)
		}
		return nil
	}
	return retry.Retry(resultCtx, writeResult, retry.WithMaxAttempts(3), retry.WithSleep(time.Second))
}

func (j *jobLoop) execJobStep(ctx context.Context, l *zap.Logger, logProvider *log.LoggerProvider, step *executeJobStep, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	// Attach pkgctx.ContextField(ctx) to BOTH the local `l` and the
	// ctx-stored logger so the otelzap bridge can extract the per-step span
	// (opened in executeJob) on every emit. The local `l` is used directly by
	// l.Info("step was completed successfully", …) below — without this the
	// caller's plain logger has no ctx field and those step-scope logs land
	// in otel_log_records with an empty span_id, which breaks the dashboard's
	// span→logs cross-link. SetLoggerWithSpan only mutates the copy stored
	// in ctx, so we have to re-wrap `l` here too.
	l = l.With(zap.String("runner_job_execution_step.name", step.name), pkgctx.ContextField(ctx))
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
		description := fmt.Sprintf("panic in %s: %s", step.name, recovered.String())
		panicErr := fmt.Errorf("panic in %s: %v", step.name, recovered.Value)

		l.Error("panic in "+step.name, zap.Error(panicErr))
		l.Error(string(debug.Stack()))

		if resultErr := j.writeFallbackJobExecutionResult(ctx, job, jobExecution, step.handler.Name(), step.name, panicErr); resultErr != nil {
			j.errRecorder.Record("write fallback job execution result", resultErr)
		}
		if updateErr := j.updateJobExecutionStatusWithDescription(ctx, job.ID, jobExecution.ID, status, description); updateErr != nil {
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
	if resultErr := j.writeFallbackJobExecutionResult(ctx, job, jobExecution, step.handler.Name(), step.name, err); resultErr != nil {
		j.errRecorder.Record("write fallback job execution result", resultErr)
	}

	// handle the error by cleaning up the execution using the handler.
	status := j.errToStatus(err)
	description := fmt.Sprintf("%s: %s", step.name, err.Error())
	if updateErr := j.updateJobExecutionStatusWithDescription(ctx, job.ID, jobExecution.ID, status, description); updateErr != nil {
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
