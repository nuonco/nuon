package syncer

import (
	"context"
	"errors"
	"time"

	configsync "github.com/nuonco/nuon/pkg/config/sync"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	attempts metric.Int64Counter
	duration metric.Float64Histogram
}

func NewMetrics(provider metric.MeterProvider) *Metrics {
	if provider == nil {
		return nil
	}
	meter := provider.Meter("github.com/nuonco/nuon/ctl-api/app-config-sync")
	attempts, err := meter.Int64Counter("nuon.app.config.sync.attempts", metric.WithUnit("{attempt}"), metric.WithDescription("Completed stored app-config sync invocations, including deferred queue provisioning."))
	if err != nil {
		return nil
	}
	duration, err := meter.Float64Histogram("nuon.app.config.sync.duration", metric.WithUnit("s"), metric.WithDescription("Elapsed time for a stored app-config sync invocation."), metric.WithExplicitBucketBoundaries(.1, .5, 1, 5, 15, 30, 60, 120, 300, 600))
	if err != nil {
		return nil
	}
	return &Metrics{attempts: attempts, duration: duration}
}

func (m *Metrics) record(ctx context.Context, start time.Time, stage string, err error) string {
	outcome := "success"
	if err != nil {
		var rejected configsync.SyncErr
		switch {
		case ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
			outcome = "cancelled"
		case errors.As(err, &rejected):
			outcome = "rejected"
		default:
			outcome = "error"
		}
	} else {
		stage = "none"
	}
	if m != nil {
		m.attempts.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome), attribute.String("stage", stage)))
		m.duration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attribute.String("outcome", outcome)))
	}
	return outcome
}
