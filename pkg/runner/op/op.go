package op

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
)

const (
	tracerName = "github.com/nuonco/nuon/bins/runner/op"
)

type EndFunc func(err error)

func Start(ctx context.Context, tool, operation string, attrs ...attribute.KeyValue) (context.Context, EndFunc) {
	tracer := pkgctx.TracerProvider(ctx).Tracer(tracerName)

	spanAttrs := make([]attribute.KeyValue, 0, len(attrs)+10)
	spanAttrs = append(spanAttrs,
		attribute.String("nuon.tool", tool),
		attribute.String("nuon.op", operation),
	)
	if meta, ok := pkgctx.GetJobMetadata(ctx); ok {
		for key, value := range map[string]string{
			"runner_job.id":                  meta.RunnerJobID,
			"runner_job_execution.id":        meta.RunnerJobExecutionID,
			"runner_job_execution_step.name": meta.StepName,
			"runner_job.group":               meta.JobGroup,
			"runner_job.operation":           meta.JobOperation,
			"runner_job.executor":            meta.Executor,
			"org.id":                         meta.OrgID,
			"install.id":                     meta.InstallID,
		} {
			if value != "" {
				spanAttrs = append(spanAttrs, attribute.String(key, value))
			}
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

func Run(ctx context.Context, tool, operation string, fn func(ctx context.Context) error, attrs ...attribute.KeyValue) error {
	ctx, end := Start(ctx, tool, operation, attrs...)
	err := fn(ctx)
	end(err)
	return err
}

func Tool(ctx context.Context, tool, operation string, attrs ...attribute.KeyValue) (context.Context, EndFunc) {
	return Start(ctx, tool, operation, attrs...)
}

func Runner(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, EndFunc) {
	return Start(ctx, "runner", operation, attrs...)
}
