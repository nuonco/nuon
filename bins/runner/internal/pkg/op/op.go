// Package op provides the primitive used to wrap individual tool / SDK call
// sites so the runner can build a tool-call graph for each job execution.
//
// Each Start call opens a span as a child of whatever span is already on the
// context (typically the per-step span opened in jobloop, which is itself a
// child of the per-execution root span). The returned context has:
//
//  1. The span installed via otel trace context (so any nested op.Start uses
//     it as a parent automatically).
//  2. A logger derived from the contextual logger that injects the *Go
//     context* itself as a zap field. The otelzap bridge inspects fields and,
//     when one carries a context.Context, uses it as the emit context. That is
//     the documented mechanism for promoting trace_id / span_id into the
//     dedicated OTLP log record fields (otel_log_records.trace_id /
//     otel_log_records.span_id) instead of the Map fallback.
//
// Stamp every span with runner_job_id / runner_job_execution_id / step name
// from pkgctx.JobMetadata: OTEL span attributes do not inherit from parents,
// so the CH ingest path can only populate the dedicated columns when each
// span carries them directly.
package op

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
)

const (
	tracerName = "github.com/nuonco/nuon/bins/runner/op"
)

// EndFunc finalizes the span. Pass the final error (or nil); status code is
// set to Error iff err != nil so dashboards can colorize failed nodes.
type EndFunc func(err error)

// Start opens a child span named "<tool>.<operation>" and returns a context
// that carries (a) the span and (b) a logger whose emit context is the same
// context — see package doc for why that matters.
//
// Callers should always defer the returned EndFunc with the final error:
//
//	ctx, end := op.Start(ctx, "terraform", "plan")
//	defer end(err)
func Start(ctx context.Context, tool, operation string, attrs ...attribute.KeyValue) (context.Context, EndFunc) {
	// Resolve the TracerProvider via ctx so we pick up the process-scoped
	// provider stamped by jobloop.executeJob; otel.Tracer would route through
	// the global, which transitive deps (e.g. docker distribution) overwrite.
	tracer := pkgctx.TracerProvider(ctx).Tracer(tracerName)

	spanAttrs := make([]attribute.KeyValue, 0, len(attrs)+5)
	spanAttrs = append(spanAttrs,
		attribute.String("nuon.tool", tool),
		attribute.String("nuon.op", operation),
	)
	if meta, ok := pkgctx.GetJobMetadata(ctx); ok {
		if meta.RunnerJobID != "" {
			spanAttrs = append(spanAttrs, attribute.String("runner_job.id", meta.RunnerJobID))
		}
		if meta.RunnerJobExecutionID != "" {
			spanAttrs = append(spanAttrs, attribute.String("runner_job_execution.id", meta.RunnerJobExecutionID))
		}
		if meta.StepName != "" {
			spanAttrs = append(spanAttrs, attribute.String("runner_job_execution_step.name", meta.StepName))
		}
	}
	spanAttrs = append(spanAttrs, attrs...)

	ctx, span := tracer.Start(ctx, tool+"."+operation,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(spanAttrs...),
	)

	if l, err := pkgctx.Logger(ctx); err == nil && l != nil {
		l = l.With(
			zap.String("nuon.tool", tool),
			zap.String("nuon.op", operation),
		)
		// SetLoggerWithSpan attaches pkgctx.ContextField(ctx) so the otelzap
		// bridge picks up this op's span on every emit through the contextual
		// logger. Without it, log records from inside this scope land with an
		// empty span_id and the dashboard's span→logs cross-link finds no
		// matches.
		ctx = pkgctx.SetLoggerWithSpan(ctx, l)
	}

	return ctx, func(err error) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}

// Run is sugar around Start for the common case where the caller does not
// need the wrapped context for anything other than the closure body.
func Run(ctx context.Context, tool, operation string, fn func(ctx context.Context) error, attrs ...attribute.KeyValue) error {
	ctx, end := Start(ctx, tool, operation, attrs...)
	err := fn(ctx)
	end(err)
	return err
}

// Tool wraps Start for tool callsites (terraform, helm, kubernetes_manifest,
// job, git, oci, docker, action). Functionally identical to Start; use it at
// tool sites so the intent reads at the callsite and stays symmetric with
// Runner / Cloud below.
func Tool(ctx context.Context, tool, operation string, attrs ...attribute.KeyValue) (context.Context, EndFunc) {
	return Start(ctx, tool, operation, attrs...)
}

// Runner wraps Start for runner-internal operations (jobloop helpers, calls
// into the ctl-api runner client, monitor goroutines). nuon.tool is fixed to
// "runner" so every runner-internal span lands under one filter value in the
// dashboard.
func Runner(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, EndFunc) {
	return Start(ctx, "runner", operation, attrs...)
}
