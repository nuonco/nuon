package exporter

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

type Config struct {
	confighttp.ClientConfig `mapstructure:",squash"`

	QueueConfig exporterhelper.QueueBatchConfig `mapstructure:"sending_queue"`
	RetryConfig configretry.BackOffConfig       `mapstructure:"retry_on_failure"`
	exporterhelper.Option
}

var _ component.Config = (*Config)(nil)

func (cfg *Config) Validate() error {
	return nil
}
