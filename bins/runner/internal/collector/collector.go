package collector

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"

	"github.com/powertoolsdev/mono/bins/runner/internal/collector/components"
)

func NewSettings() otelcol.CollectorSettings {
	info := component.BuildInfo{
		Command:     "nuon-runner-otelcol",
		Description: "Nuon Runner Custom OTEL Collector",
		Version:     "1.0.0",
	}

	return otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: components.All,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				ProviderFactories: []confmap.ProviderFactory{
					envprovider.NewFactory(),
					fileprovider.NewFactory(),
					httpprovider.NewFactory(),
					httpsprovider.NewFactory(),
					yamlprovider.NewFactory(),
				},
				ConverterFactories: []confmap.ConverterFactory{
					expandconverter.NewFactory(),
				},
			},
		},
	}
}

func New() (*otelcol.Collector, error) {
	set := NewSettings()
	col, err := otelcol.NewCollector(set)
	if err != nil {
		return nil, fmt.Errorf("unable to create collector: %w", err)
	}

	return col, err
}
