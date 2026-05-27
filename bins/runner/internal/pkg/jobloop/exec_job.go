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

	"github.com/nuonco/nuon/bins/runner/internal/jobs"
	"github.com/nuonco/nuon/bins/runner/internal/jobs/sandboxhandler"
	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/errs"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/log"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/slog"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/workspace"
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
	l = l.With(zap.String("runner_job.type", string(job.Type)))
	l = l.With(zap.String("log_stream.id", job.LogStreamID))

	defer func() {
		if err := jl.ForceFlush(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			l.Error("unable to flush logger", zap.Error(err))
		}
	}()

	// create an execution in the API
	l.Info("creating job execution")
	execution, err := j.apiClient.CreateJobExecution(ctx, job.ID, new(models.ServiceCreateRunnerJobExecutionRequest))
	if err != nil {
		return errors.Wrap(err, "unable to create execution")
	}
	l = l.With(zap.String("runner_job_execution.id", execution.ID))

	// Open the per-execution root span. Every step / op span is a descendant
	// so the entire job execution forms a single trace. Job metadata goes onto
	// ctx so op.Start can stamp it on every descendant span without each
	// callsite having to repeat itself.
	ctx = pkgctx.SetJobMetadata(ctx, pkgctx.JobMetadata{
		RunnerJobID:          job.ID,
		RunnerJobExecutionID: execution.ID,
	})
	// Stash the process-scoped TracerProvider into ctx so op.Start sees it
	// and we don't get poisoned by transitive deps that overwrite the OTEL
	// global (notably the docker distribution registry).
	tp := j.processRegistrar.TracerProvider()
	ctx = pkgctx.SetTracerProvider(ctx, tp)
	tracer := tp.Tracer("github.com/nuonco/nuon/bins/runner/jobloop")
	ctx, rootSpan := tracer.Start(ctx, "job."+string(job.Type),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("nuon.tool", "runner"),
			attribute.String("nuon.job.type", string(job.Type)),
			attribute.String("nuon.job.operation", string(job.Operation)),
			attribute.String("runner_job.id", job.ID),
			attribute.String("runner_job_execution.id", execution.ID),
		),
	)
	// Re-wrap `l` with pkgctx.ContextField(ctx) so the otelzap bridge can
	// extract the rootSpan on every emit. Without this, every "creating job
	// execution" / "getting job handler" / "finished job" log lands in
	// otel_log_records with an empty span_id and the dashboard's span→logs
	// cross-link finds no matches when the user clicks the job span.
	l = l.With(pkgctx.ContextField(ctx))
	ctx = pkgctx.SetLogger(ctx, l)
	var jobErr error
	defer func() {
		if jobErr != nil {
			rootSpan.RecordError(jobErr)
			rootSpan.SetStatus(codes.Error, jobErr.Error())
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

		return err
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
		stepCtx := pkgctx.SetJobMetadata(ctx, pkgctx.JobMetadata{
			RunnerJobID:          job.ID,
			RunnerJobExecutionID: execution.ID,
			StepName:             step.name,
		})
		stepCtx, stepSpan := tracer.Start(stepCtx, "step."+step.name,
			trace.WithSpanKind(trace.SpanKindInternal),
			trace.WithAttributes(
				attribute.String("nuon.tool", "runner"),
				attribute.String("runner_job_execution_step.name", step.name),
				attribute.String("runner_job.id", job.ID),
				attribute.String("runner_job_execution.id", execution.ID),
			),
		)
		// Step-scope logger picks up the step span via ContextField so the
		// "executing job step …" marker lands on the step span instead of
		// the parent rootSpan.
		stepL := l.With(pkgctx.ContextField(stepCtx))
		stepL.Info("executing job step "+step.name, zap.String("step", step.name))
		stepErr := j.execJobStep(stepCtx, stepL, jl, step, job, execution)
		if stepErr != nil {
			stepSpan.RecordError(stepErr)
			stepSpan.SetStatus(codes.Error, stepErr.Error())
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
