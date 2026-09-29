package jobloop

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/jobs/sandboxhandler"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/audit"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/slog"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/errcapture"
	"github.com/nuonco/nuon/pkg/runner/errs"
	"github.com/nuonco/nuon/pkg/runner/jobs"
	"github.com/nuonco/nuon/pkg/runner/log"
	"github.com/nuonco/nuon/pkg/runner/workspace"
)

type executeJobStep struct {
	name      string
	fn        func(context.Context, jobs.JobHandler, *models.AppRunnerJob, *models.AppRunnerJobExecution) error
	cleanupFn func(context.Context, jobs.JobHandler, *models.AppRunnerJob, *models.AppRunnerJobExecution) error
	handler   jobs.JobHandler

	startStatus models.AppRunnerJobExecutionStatus
}

func (j *jobLoop) executeJob(ctx context.Context, job *models.AppRunnerJob) error {
	job.RunnerProcessID = j.processRegistrar.ProcessID()

	jl, err := slog.NewOTELProvider(j.cfg, j.settings, job.LogStreamID)
	if err != nil {
		return errors.Wrap(err, "unable to create otel provider")
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := jl.Shutdown(shutdownCtx); err != nil {
			j.l.Error("unable to shut down job logger", zap.Error(err))
		}
	}()

	l, err := log.NewOTELJobLogger(j.cfg, jl)
	if err != nil {
		return errors.Wrap(err, "unable to get job logger")
	}

	l = l.With(zap.String("runner_job.id", job.ID))
	l = l.With(zap.String("runner_job.type", string(job.Type)))
	l = l.With(zap.String("log_stream.id", job.LogStreamID))

	l.Info("creating job execution")
	execution, err := j.apiClient.CreateJobExecution(ctx, job.ID, new(models.ServiceCreateRunnerJobExecutionRequest))
	if err != nil {
		return errors.Wrap(err, "unable to create execution")
	}
	l = l.With(zap.String("runner_job_execution.id", execution.ID))

	// why: Per-execution status coalescer. Intermediate status transitions
	// (resetting → fetching → validate → initialize → ...) now drop
	// non-terminal pings on the floor while the previous write is in
	// flight; terminal statuses still land synchronously and in order.
	// The deferred Close() is a guard against panics — WriteTerminal
	// is idempotent (stopOnce) so the normal path is unaffected.
	coalescer := newStatusCoalescer(job.ID, execution.ID, l, j.writeJobExecutionStatus)
	j.attachCoalescer(execution.ID, coalescer)
	defer func() {
		coalescer.Close()
		j.detachCoalescer(execution.ID)
	}()

	ctx = pkgctx.SetJobMetadata(ctx, jobs.AuditMetadata(job, execution.ID, ""))
	// why: Stash the process-scoped TracerProvider into ctx so op.Start sees it
	// and we don't get poisoned by transitive deps that overwrite the OTEL
	// global (notably the docker distribution registry).
	tp := j.processRegistrar.TracerProvider()
	ctx = pkgctx.SetTracerProvider(ctx, tp)
	tracer := tp.Tracer("github.com/nuonco/nuon/bins/runner/jobloop")
	rootSpanAttrs := append([]attribute.KeyValue{
		attribute.String("nuon.tool", "runner"),
		attribute.String("nuon.job.type", string(job.Type)),
		attribute.String("nuon.job.operation", string(job.Operation)),
		attribute.String("runner_job.id", job.ID),
		attribute.String("runner_job_execution.id", execution.ID),
	}, jobs.AuditAttrs(job)...)
	ctx, rootSpan := tracer.Start(ctx, "job."+string(job.Type),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(rootSpanAttrs...),
	)
	l = l.With(pkgctx.ContextField(ctx))

	capture := errcapture.New()
	l = l.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, capture.Core())
	}))
	ctx = errcapture.NewContext(ctx, capture)

	ctx = pkgctx.SetLogger(ctx, l)

	j.writeJobAudit(job, "job execution started", audit.OutcomeStarted, nil)

	var jobErr error
	defer func() {
		if jobErr != nil {
			j.writeJobAudit(job, "job execution failed", audit.OutcomeFailed, map[string]string{"error": jobErr.Error()})
			rootSpan.RecordError(jobErr)
			rootSpan.SetStatus(codes.Error, jobErr.Error())
		} else {
			j.writeJobAudit(job, "job execution finished", audit.OutcomeSucceeded, nil)
		}
		rootSpan.End()
	}()

	defer workspace.CleanupByID(execution.ID)

	l.Info("getting job handler")
	handler, err := j.getHandler(job)
	if err != nil {
		l.Error("no valid job handler found for job",
			zap.String("type", string(job.Type)),
			zap.Error(err),
		)
		description := fmt.Sprintf("no valid job handler for job type %s: %s", job.Type, err.Error())
		if resultErr := j.writeFallbackJobExecutionResult(ctx, job, execution, "jobloop", "get-handler", err); resultErr != nil {
			j.errRecorder.Record("write fallback job execution result", resultErr)
		}
		if updateErr := j.updateJobExecutionStatusWithDescription(ctx, job.ID, execution.ID, models.AppRunnerJobExecutionStatusFailed, description); updateErr != nil {
			j.errRecorder.Record("no handler found", updateErr)
		}

		jobErr = err
		return jobErr
	}

	if j.isSandbox(job) {
		l.Info("sandbox mode active, replacing handler with sandbox handler",
			zap.String("job_type", string(job.Type)),
			zap.String("operation", string(job.Operation)),
			zap.String("job_id", job.ID),
			zap.Bool("sandbox_mode_setting", j.settings.SandboxMode),
		)

		var sandboxCfg *sandboxhandler.Config
		apiCfg, err := j.apiClient.GetSandboxConfig(ctx, string(job.Type), string(job.Operation))
		if err != nil {
			l.Warn("unable to fetch sandbox config from API, using defaults",
				zap.Error(err))
		}
		if apiCfg != nil {
			sandboxCfg = sandboxhandler.ConfigFromAPI(apiCfg)
		}

		handler = sandboxhandler.New(sandboxCfg, j.apiClient, j.cfg, j.shutdowner, job, execution)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, time.Duration(job.ExecutionTimeout))
	defer cancel()

	doneCh := make(chan struct{})
	defer close(doneCh)
	go func() {
		j.monitorJob(ctx, cancel, doneCh, job.ID, l, handler)
	}()

	steps, err := j.getJobSteps(ctx, handler)
	if err != nil {
		jobErr = errors.Wrap(err, "unable to get job steps")
		return jobErr
	}

	for _, step := range steps {
		stepCtx := pkgctx.SetJobMetadata(ctx, jobs.AuditMetadata(job, execution.ID, step.name))
		stepSpanAttrs := append([]attribute.KeyValue{
			attribute.String("nuon.tool", "runner"),
			attribute.String("runner_job_execution_step.name", step.name),
			attribute.String("runner_job.id", job.ID),
			attribute.String("runner_job_execution.id", execution.ID),
		}, jobs.AuditAttrs(job)...)
		stepCtx, stepSpan := tracer.Start(stepCtx, "step."+step.name,
			trace.WithSpanKind(trace.SpanKindInternal),
			trace.WithAttributes(stepSpanAttrs...),
		)
		stepL := l.With(pkgctx.ContextField(stepCtx))
		stepL.Info("executing job step "+step.name, zap.String("step", step.name))

		stepAttrs := map[string]string{"runner_job_execution_step.name": step.name}
		j.writeJobAudit(job, "job step started", audit.OutcomeStarted, stepAttrs)

		stepErr := j.execJobStep(stepCtx, stepL, jl, step, job, execution)
		if stepErr != nil {
			stepAttrs["error"] = stepErr.Error()
			j.writeJobAudit(job, "job step failed", audit.OutcomeFailed, stepAttrs)
			stepSpan.RecordError(stepErr)
			stepSpan.SetStatus(codes.Error, stepErr.Error())
		} else {
			j.writeJobAudit(job, "job step finished", audit.OutcomeSucceeded, stepAttrs)
		}
		stepSpan.End()
		if stepErr != nil {
			jobErr = errs.WithHandlerError(stepErr, j.jobGroup, step.name, job.Type)
			return jobErr
		}
	}

	if err := j.updateJobExecutionStatus(ctx, job.ID, execution.ID, models.AppRunnerJobExecutionStatusFinished); err != nil {
		jobErr = errors.Wrap(err, "unable to update job execution status after successful execution")
		return jobErr
	}

	l.Info("finished job", zap.String("name", handler.Name()))

	return nil
}

func (j *jobLoop) writeJobAudit(job *models.AppRunnerJob, message, outcome string, attributes map[string]string) {
	event, ok := audit.JobEvent(job, message, outcome, attributes)
	if !ok {
		return
	}
	if err := j.audit.WriteAsync(event); err != nil {
		if errors.Is(err, audit.ErrUnavailable) {
			return
		}
		j.l.Warn("customer job audit event enqueue failed",
			zap.String("audit.event", event.Name),
			zap.String("audit.outcome", event.Outcome),
			zap.Error(err),
		)
	}
}
