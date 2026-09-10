package telemetry

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.uber.org/fx"
)

func NewMeterProvider(lc fx.Lifecycle, cfg *Config) (metric.MeterProvider, error) {
	if cfg.Endpoint == "" {
		return noop.NewMeterProvider(), nil
	}

	endpoint, err := url.JoinPath(cfg.Endpoint, "v1/metrics")
	if err != nil {
		return nil, fmt.Errorf("configure OTEL metric endpoint: %w", err)
	}
	exporter, err := otlpmetrichttp.New(context.Background(),
		otlpmetrichttp.WithEndpointURL(endpoint),
		otlpmetrichttp.WithTemporalitySelector(sdkmetric.CumulativeTemporalitySelector),
		otlpmetrichttp.WithAggregationSelector(sdkmetric.DefaultAggregationSelector),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to configure OTLP metric exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(cfg.Resource),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		sdkmetric.WithCardinalityLimit(2000),
		sdkmetric.WithExemplarFilter(exemplar.AlwaysOffFilter),
	)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return provider.Shutdown(ctx)
		},
	})
	return provider, nil
}
