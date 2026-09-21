package hooks

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	deliveryChannelWebhook  = "webhook"
	deliveryChannelSlack    = "slack"
	deliveryOperationPost   = "post"
	deliveryOperationUpdate = "update"
)

type deliveryMetrics struct {
	attempts metric.Int64Counter
	duration metric.Float64Histogram
}

func newDeliveryMetrics(provider metric.MeterProvider) *deliveryMetrics {
	if provider == nil {
		return nil
	}
	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/notification-delivery")
	attempts, err := meter.Int64Counter(
		"nuon.notification.delivery.attempts",
		metric.WithUnit("{attempt}"),
		metric.WithDescription("Outbound notification delivery attempts."),
	)
	if err != nil {
		return nil
	}
	duration, err := meter.Float64Histogram(
		"nuon.notification.delivery.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of outbound notification delivery attempts."),
		metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10),
	)
	if err != nil {
		return nil
	}
	return &deliveryMetrics{attempts: attempts, duration: duration}
}

func (m *deliveryMetrics) record(ctx context.Context, channel, operation string, started time.Time, err error) {
	if m == nil {
		return
	}
	outcome := "success"
	if err != nil {
		outcome = "failure"
	}
	attrs := metric.WithAttributes(
		attribute.String("nuon.notification.channel", channel),
		attribute.String("nuon.notification.operation", operation),
		attribute.String("nuon.notification.outcome", outcome),
	)
	m.attempts.Add(ctx, 1, attrs)
	m.duration.Record(ctx, time.Since(started).Seconds(), attrs)
}
