package metrics

import (
	"context"
	"fmt"
	"os"

	ddotel "github.com/DataDog/dd-trace-go/v2/ddtrace/opentelemetry"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"
)

func New(v *validator.Validate, l *zap.Logger, cfg *internal.Config) (metrics.Writer, error) {
	tags := []string{
		fmt.Sprintf("service_deployment:%s", cfg.ServiceDeployment),
		fmt.Sprintf("service_type:%s", cfg.ServiceType),
	}
	tags = append(tags, cfg.MetricsTags...)

	mw, err := metrics.New(v,
		metrics.WithDisable(cfg.DisableMetrics),
		metrics.WithTags(tags...),
		metrics.WithLogger(l),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create new metrics writer: %w", err)
	}

	// Local-dev affordances mirroring the metric debug log lines: dump spans to a JSON-lines
	// file and/or a local OTLP viewer (e.g. otel-desktop-viewer, Jaeger) so traces can be
	// inspected without a Datadog agent.
	var localExporters []sdktrace.TracerProviderOption
	if traceFile := os.Getenv("OTEL_LOCAL_TRACE_FILE"); traceFile != "" {
		f, err := os.OpenFile(traceFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("unable to open local trace file: %w", err)
		}
		exp, err := stdouttrace.New(stdouttrace.WithWriter(f))
		if err != nil {
			return nil, fmt.Errorf("unable to create local trace exporter: %w", err)
		}
		localExporters = append(localExporters, sdktrace.WithSyncer(exp))
		l.Info("local OTel trace export enabled", zap.String("file", traceFile))
	}
	if otlpEndpoint := os.Getenv("OTEL_LOCAL_TRACE_OTLP"); otlpEndpoint != "" {
		exp, err := otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpoint(otlpEndpoint),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create local OTLP trace exporter: %w", err)
		}
		localExporters = append(localExporters, sdktrace.WithBatcher(exp))
		l.Info("local OTel trace export enabled", zap.String("otlp_endpoint", otlpEndpoint))
	}
	if len(localExporters) > 0 {
		opts := append(localExporters, sdktrace.WithResource(resource.NewSchemaless(
			semconv.ServiceName("ctl-api"),
			attribute.String("service_deployment", cfg.ServiceDeployment),
		)))
		otel.SetTracerProvider(sdktrace.NewTracerProvider(opts...))
		return mw, nil
	}

	if !cfg.DisableMetrics {
		// The ddotel bridge starts the Datadog tracer internally and exposes it behind the
		// vendor-neutral OTel API, so instrumentation (e.g. the gorm tracing plugin) can later be
		// pointed at an OTLP backend by swapping the provider without re-instrumenting.
		// The node-level agent listens on hostPort 8126 for APM (apm.portEnabled), so the
		// trace agent address must target HOST_IP like dogstatsd does; the default
		// localhost:8126 would resolve to the pod itself and traces would be dropped.
		tp := ddotel.NewTracerProvider(
			tracer.WithRuntimeMetrics(),
			tracer.WithDogstatsdAddr(fmt.Sprintf("%s:8125", os.Getenv("HOST_IP"))),
			tracer.WithAgentAddr(fmt.Sprintf("%s:8126", os.Getenv("HOST_IP"))),
		)
		otel.SetTracerProvider(tp)
	}

	return mw, nil
}
