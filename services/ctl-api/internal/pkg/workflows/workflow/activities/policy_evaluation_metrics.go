package activities

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type policyEvaluationMetrics struct {
	count    metric.Int64Counter
	duration metric.Float64Histogram
}

func newPolicyEvaluationMetrics(provider metric.MeterProvider) policyEvaluationMetrics {
	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/workflows/policy-evaluation")
	count, _ := meter.Int64Counter("nuon.policy.evaluation.count",
		metric.WithUnit("{evaluation}"),
		metric.WithDescription("Policy-input evaluations by outcome, decision, or bounded error stage."))
	duration, _ := meter.Float64Histogram("nuon.policy.evaluation.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of a policy-input evaluation by outcome, decision, or bounded error stage."),
		metric.WithExplicitBucketBoundaries(.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60, 90))
	return policyEvaluationMetrics{count: count, duration: duration}
}

func (m policyEvaluationMetrics) record(ctx context.Context, started time.Time, decision, errorStage string) {
	if m.count == nil || m.duration == nil {
		return
	}
	attrs := []attribute.KeyValue{attribute.String("outcome", "success")}
	if errorStage != "" {
		attrs[0] = attribute.String("outcome", "error")
		attrs = append(attrs, attribute.String("error.type", errorStage))
	} else {
		attrs = append(attrs, attribute.String("decision", decision))
	}
	options := metric.WithAttributes(attrs...)
	m.count.Add(ctx, 1, options)
	m.duration.Record(ctx, time.Since(started).Seconds(), options)
}
