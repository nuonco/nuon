package state

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	operations metric.Int64Counter
	duration   metric.Float64Histogram
}

func NewMetrics(provider metric.MeterProvider) *Metrics {
	if provider == nil {
		return nil
	}
	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/install-state")
	operations, err := meter.Int64Counter("nuon.install.state.operations", metric.WithUnit("{operation}"))
	if err != nil {
		return nil
	}
	duration, err := meter.Float64Histogram("nuon.install.state.operation.duration", metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60))
	if err != nil {
		return nil
	}
	return &Metrics{operations: operations, duration: duration}
}

func (m *Metrics) Record(ctx context.Context, operation string, started time.Time, err error) {
	if m == nil {
		return
	}
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	attrs := metric.WithAttributes(attribute.String("operation", operation), attribute.String("outcome", outcome))
	m.operations.Add(ctx, 1, attrs)
	m.duration.Record(ctx, time.Since(started).Seconds(), attrs)
}
