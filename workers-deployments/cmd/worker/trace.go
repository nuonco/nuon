package worker

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTraceGrpcPort int = 4317
)

func (w *worker) getTracer() (trace.Tracer, error) {
	otlpExporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", w.Config.HostIP, defaultTraceGrpcPort)))
	if err != nil {
		return nil, fmt.Errorf("unable to create otlptrace exporter: %w", err)
	}

	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(otlpExporter),
	)

	otel.SetTracerProvider(tracerProvider)
	return tracerProvider.Tracer("nuon.workers-deployments"), nil
}
