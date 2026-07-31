package jobloop

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon/bins/runner/internal/jobs/sandboxhandler"
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

	// create an execution in the API
	l.Info("creating job execution")
	execution, err := j.apiClient.CreateJobExecution(ctx, job.ID, new(models.ServiceCreateRunnerJobExecutionRequest))
	if err != nil {
		return errors.Wrap(err, "unable to create execution")
	}
	l = l.With(zap.String("runner_job_execution.id", execution.ID))

	// Per-execution status coalescer. Intermediate status transitions
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

	// Open the per-execution root span. Every step / op span is a descendant
	// so the entire job execution forms a single trace. Job metadata goes onto
	// ctx so op.Start can stamp it on every descendant span without each
	// callsite having to repeat itself.
	ctx = pkgctx.SetJobMetadata(ctx, jobs.AuditMetadata(job, execution.ID, ""))
	// Stash the process-scoped TracerProvider into ctx so op.Start sees it
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
	// Re-wrap `l` with pkgctx.ContextField(ctx) so the otelzap bridge can
	// extract the rootSpan on every emit. Without this, every "creating job
	// execution" / "getting job handler" / "finished job" log lands in
	// otel_log_records with an empty span_id and the dashboard's span→logs
	// cross-link finds no matches when the user clicks the job span.
	l = l.With(pkgctx.ContextField(ctx))

	// Tee an error-capture core into the job logger so every error-level record
	// (including terraform's structured @level:"error" diagnostics) is buffered
	// for this execution. The API client decorator attaches the buffer to a
	// failed result so ctl-api parses the real cause, not the thin wrapper.
	capture := errcapture.New()
	l = l.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(c, capture.Core())
	}))
	ctx = errcapture.NewContext(ctx, capture)

	ctx = pkgctx.SetLogger(ctx, l)

	auditL := log.NewAudit(l, job)
	log.AuditEvent(auditL, "job execution started", log.OutcomeStarted)

	var jobErr error
	defer func() {
		if jobErr != nil {
			log.AuditEvent(auditL, "job execution failed", log.OutcomeFailed, zap.Error(jobErr))
			rootSpan.RecordError(jobErr)
			rootSpan.SetStatus(codes.Error, jobErr.Error())
		} else {
			log.AuditEvent(auditL, "job execution finished", log.OutcomeSucceeded)
		}
		rootSpan.End()
	}()

	// Always clean up the workspace directory for this execution, even if
	// the job panics or errors before the cleanup step runs. This uses the
	// workspace package directly so it works regardless of handler state.
	defer workspace.CleanupByID(execution.ID)

	l.Info("getting job handler")
	handler, err := j.getHandler(job)
	if err != nil {
		l.Error("no valid job handler found for job",
			zap.String("type", string(job.Type)),
			zap.Error(err),
		)
		description := fmt.Sprintf("no valid job handler for job type %s: %s", job.Type, err.Error())
		if updateErr := j.updateJobExecutionStatusWithDescription(ctx, job.ID, execution.ID, models.AppRunnerJobExecutionStatusFailed, description); updateErr != nil {
			j.errRecorder.Record("no handler found", updateErr)
		}

		jobErr = err
		return jobErr
	}

	// If sandbox mode, fetch config from API and replace handler
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
		// Open per-step span as a child of the execution root FIRST so the
		// "executing job step …" log we emit below carries the step span_id.
		// Stamp the step name onto JobMetadata so anything launched inside
		// the step (op.Start callsites in deploy / sandbox handlers)
		// inherits it.
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
		// Step-scope logger picks up the step span via ContextField so the
		// "executing job step …" marker lands on the step span instead of
		// the parent rootSpan.
		stepL := l.With(pkgctx.ContextField(stepCtx))
		stepL.Info("executing job step "+step.name, zap.String("step", step.name))

		stepAuditL := log.NewAudit(stepL, job)
		stepNameField := zap.String("runner_job_execution_step.name", step.name)
		log.AuditEvent(stepAuditL, "job step started", log.OutcomeStarted, stepNameField)

		stepErr := j.execJobStep(stepCtx, stepL, jl, step, job, execution)
		if stepErr != nil {
			log.AuditEvent(stepAuditL, "job step failed", log.OutcomeFailed, stepNameField, zap.Error(stepErr))
			stepSpan.RecordError(stepErr)
			stepSpan.SetStatus(codes.Error, stepErr.Error())
		} else {
			log.AuditEvent(stepAuditL, "job step finished", log.OutcomeSucceeded, stepNameField)
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
