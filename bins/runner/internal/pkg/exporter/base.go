package exporter

import (
	"context"
	"fmt"

	nuonrunner "github.com/nuonco/nuon-runner-go"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.uber.org/zap"
)

type baseExporter struct {
	// Input  configuration.
	config   *Config
	logger   *zap.Logger
	settings component.TelemetrySettings

	apiClient nuonrunner.Client

	// config settings
	APIURL    string
	RunnerID  string
	AuthToken string
}

func (e *baseExporter) start(ctx context.Context, host component.Host) error {
	apiClient, err := nuonrunner.New(
		nuonrunner.WithURL(e.APIURL),
		nuonrunner.WithRunnerID(e.RunnerID),
		nuonrunner.WithAuthToken(e.AuthToken),
	)
	if err != nil {
		return fmt.Errorf("unable to create runner api client: %w", err)
	}

	e.apiClient = apiClient

	return nil
}

func newExporter(cfg component.Config, set exporter.Settings) (*baseExporter, error) {
	oCfg := cfg.(*Config)

	// client construction is deferred to start
	return &baseExporter{
		config:   oCfg,
		logger:   set.Logger,
		settings: set.TelemetrySettings,
	}, nil
}
