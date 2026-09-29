package pkgctx

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var loggerNotFoundErr error = fmt.Errorf("logger not found in context")

type logCtxKey struct{}

func Logger(ctx context.Context) (*zap.Logger, error) {
	val := ctx.Value(logCtxKey{})
	if val == nil {
		return nil, loggerNotFoundErr
	}

	return val.(*zap.Logger), nil
}

func SetLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, logCtxKey{}, l)
}

func LoggerOrDefault(ctx context.Context, def *zap.Logger) *zap.Logger {
	if l, err := Logger(ctx); err == nil && l != nil {
		return l
	}
	return def
}

const zapCtxFldName = "ctx"

// why: ContextField builds a zap field whose Interface is the Go context. otelzap
// detects the context.Context value and uses it as the emit context, which is
// what makes trace_id / span_id land in the dedicated OTLP log record columns
// rather than as Map attributes. We use SkipType so the dev / stdout encoders
// do not try to print the context.
func ContextField(ctx context.Context) zap.Field {
	return zap.Field{
		Key:       zapCtxFldName,
		Type:      zapcore.SkipType,
		Interface: ctx,
	}
}

// why: SetLoggerWithSpan is the same as SetLogger but also decorates the logger
// with ContextField(ctx) so the otelzap bridge can extract span_id / trace_id
// from the current span on every emit. Use this whenever you store a logger
// into a context that carries a new span (jobloop step boundary, op.Start,
// per-handler logger setup) — without it, downstream log calls land in
// otel_log_records with empty span_id even though the trace tab has spans.
func SetLoggerWithSpan(ctx context.Context, l *zap.Logger) context.Context {
	l = l.With(ContextField(ctx))
	return SetLogger(ctx, l)
}

type JobMetadata struct {
	RunnerJobID          string
	RunnerJobExecutionID string
	StepName             string

	JobGroup     string
	JobOperation string
	Executor     string
	OrgID        string
	InstallID    string
}

type jobMetaCtxKey struct{}

func SetJobMetadata(ctx context.Context, meta JobMetadata) context.Context {
	return context.WithValue(ctx, jobMetaCtxKey{}, meta)
}

func GetJobMetadata(ctx context.Context) (JobMetadata, bool) {
	v := ctx.Value(jobMetaCtxKey{})
	if v == nil {
		return JobMetadata{}, false
	}
	m, ok := v.(JobMetadata)
	return m, ok
}

type tracerProviderCtxKey struct{}

func SetTracerProvider(ctx context.Context, tp oteltrace.TracerProvider) context.Context {
	return context.WithValue(ctx, tracerProviderCtxKey{}, tp)
}

func TracerProvider(ctx context.Context) oteltrace.TracerProvider {
	if v := ctx.Value(tracerProviderCtxKey{}); v != nil {
		if tp, ok := v.(oteltrace.TracerProvider); ok && tp != nil {
			return tp
		}
	}
	return otel.GetTracerProvider()
}
