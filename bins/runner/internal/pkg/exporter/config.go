package exporter

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines configuration for OTLP/HTTP exporter.
type Config struct {
	confighttp.ClientConfig `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.

	QueueConfig exporterhelper.QueueBatchConfig `mapstructure:"sending_queue"`
	RetryConfig configretry.BackOffConfig       `mapstructure:"retry_on_failure"`
	exporterhelper.Option
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
