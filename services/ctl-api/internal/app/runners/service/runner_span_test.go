package service

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceRunnerRequestOperation(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previousTracer := runnerTracer
	runnerTracer = provider.Tracer("test")
	t.Cleanup(func() {
		runnerTracer = previousTracer
		require.NoError(t, provider.Shutdown(context.Background()))
	})

	rootCtx, rootSpan := provider.Tracer("test").Start(context.Background(), "request")
	responseRecorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(responseRecorder)
	ctx.Request = httptest.NewRequest("POST", "/v1/runners/id/processes", nil).WithContext(rootCtx)
	originalRequest := ctx.Request

	err := traceRunnerRequestOperation(ctx, "runner.process.test", func(operationCtx context.Context) error {
		require.Equal(t, trace.SpanContextFromContext(ctx.Request.Context()), trace.SpanContextFromContext(operationCtx))
		require.NotEqual(t, rootSpan.SpanContext(), trace.SpanContextFromContext(operationCtx))
		return nil
	})
	require.NoError(t, err)
	require.Same(t, originalRequest, ctx.Request)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, rootSpan.SpanContext().SpanID(), spans[0].Parent().SpanID())
	rootSpan.End()
}
