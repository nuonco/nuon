package activities

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type componentHealthEvaluationMetrics struct {
	attempts metric.Int64Counter
	duration metric.Float64Histogram
}

func newComponentHealthEvaluationMetrics(provider metric.MeterProvider) *componentHealthEvaluationMetrics {
	if provider == nil {
		return nil
	}
	meter := provider.Meter("github.com/nuonco/nuon/ctl-api/component-health")
	attempts, err := meter.Int64Counter("nuon.install.component.health.evaluation.attempts",
		metric.WithUnit("{attempt}"),
		metric.WithDescription("Completed install-level component health evaluation invocations by outcome and reason."))
	if err != nil {
		return nil
	}
	duration, err := meter.Float64Histogram("nuon.install.component.health.evaluation.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Elapsed time of install-level component health evaluations, excluding skipped invocations."),
		metric.WithExplicitBucketBoundaries(.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60))
	if err != nil {
		return nil
	}
	return &componentHealthEvaluationMetrics{attempts: attempts, duration: duration}
}

func (m *componentHealthEvaluationMetrics) recordForInstall(ctx context.Context, started time.Time, installID, reason string, result *EvaluateComponentHealthResponse, err error) {
	if m == nil || (result == nil && err == nil) {
		return
	}
	outcome := "success"
	switch {
	case err != nil:
		outcome = "error"
		if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			outcome = "cancelled"
		}
	case result.Skipped:
		outcome = "skipped"
	default:
		reason = "none"
	}
	attrs := []attribute.KeyValue{attribute.String("outcome", outcome), attribute.String("reason", reason)}
	if installID != "" {
		attrs = append(attrs, attribute.String("nuon.install.id", installID))
	}
	m.attempts.Add(ctx, 1, metric.WithAttributes(attrs...))
	if outcome != "skipped" {
		m.duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(attribute.String("outcome", outcome)))
	}
}
