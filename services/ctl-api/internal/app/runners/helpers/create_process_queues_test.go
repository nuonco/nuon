package helpers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTraceProcessOperation(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previousTracer := processTracer
	processTracer = provider.Tracer("test")
	t.Cleanup(func() {
		processTracer = previousTracer
		require.NoError(t, provider.Shutdown(context.Background()))
	})

	expectedErr := errors.New("operation failed")
	err := traceProcessOperation(context.Background(), "runner.process.test", func(context.Context) error {
		return expectedErr
	})
	require.ErrorIs(t, err, expectedErr)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "runner.process.test", spans[0].Name())
	require.Equal(t, codes.Error, spans[0].Status().Code)
	require.Len(t, spans[0].Events(), 1)
}
