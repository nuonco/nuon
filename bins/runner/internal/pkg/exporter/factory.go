package exporter

import (
	"time"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

var exporterType = component.MustNewType("nuon-otlphttp")

func NewFactory(apiClient nuonrunner.Client) exporter.Factory {
	return exporter.NewFactory(
		exporterType,
		createDefaultConfig,
		exporter.WithTraces(createTracesExporter, component.StabilityLevelBeta),
		exporter.WithMetrics(createMetricsExporter, component.StabilityLevelBeta),
		exporter.WithLogs(createLogsExporter, component.StabilityLevelBeta),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		RetryConfig: configretry.NewDefaultBackOffConfig(),
		QueueConfig: exporterhelper.NewDefaultQueueConfig(),
		ClientConfig: confighttp.ClientConfig{
			Endpoint: "",
			Timeout:  30 * time.Second,
			Headers:  map[string]configopaque.String{},
			// Default to gzip compression
			Compression: configcompression.TypeGzip,
			// We almost read 0 bytes, so no need to tune ReadBufferSize.
			WriteBufferSize: 512 * 1024,
		},
	}
}
