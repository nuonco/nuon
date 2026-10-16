package activities

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

func (a *Activities) recordComponentHealthEvaluation(ctx context.Context, started time.Time, install *app.Install, failedComponent *app.InstallComponent, reason string, result *EvaluateComponentHealthResponse, err error) {
	if result == nil && err == nil {
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
	elapsed := time.Since(started)
	orgID, appID := install.OrgID, install.AppID
	if reason == "load_install" || reason == "install_missing" {
		orgID, appID = "", ""
	}
	attrs := []attribute.KeyValue{attribute.String("outcome", outcome), attribute.String("reason", reason)}
	fields := []zap.Field{
		zap.String("source", "component_health_evaluator"),
		zap.String("flow_event", "install.health_evaluation_failed"),
		zap.String("outcome", outcome),
		zap.String("stage", reason),
		zap.Int64("duration_ms", elapsed.Milliseconds()),
	}
	for _, id := range []struct{ name, value string }{{"org", orgID}, {"app", appID}, {"install", install.ID}} {
		if id.value != "" {
			attrs = append(attrs, attribute.String("nuon."+id.name+".id", id.value))
			fields = append(fields, zap.String(id.name+"_id", id.value))
		}
	}
	if err != nil {
		if failedComponent != nil {
			fields = append(fields, zap.String("component_id", failedComponent.ComponentID), zap.String("install_component_id", failedComponent.ID))
		}
		a.l.Info("flow telemetry", fields...)
	}
	m := a.healthMetrics
	if m == nil {
		return
	}
	m.attempts.Add(ctx, 1, metric.WithAttributes(attrs...))
	if outcome != "skipped" {
		m.duration.Record(ctx, elapsed.Seconds(), metric.WithAttributes(attribute.String("outcome", outcome)))
	}
}

func (a *Activities) logComponentHealthChange(install *app.Install, e *componentEval) {
	fields := []zap.Field{
		zap.String("flow_event", "component.health_verdict_changed"),
		zap.String("source", "component_health_evaluator"),
		zap.String("org_id", install.OrgID),
		zap.String("app_id", install.AppID),
		zap.String("install_id", install.ID),
		zap.String("component_id", e.ic.ComponentID),
		zap.String("install_component_id", e.ic.ID),
		zap.String("health", string(e.verdict)),
		zap.String("previous_health", string(e.prior)),
		zap.String("component_status", string(e.ic.Status)),
		zap.Bool("cluster_seen", componentClusterSeen(e.ic) || e.clusterEvidence),
		zap.Bool("cluster_blind", e.clusterBlind),
	}
	if e.latest != nil {
		fields = append(fields,
			zap.String("observation_at", e.latest.ObservedAt.UTC().Format(time.RFC3339)),
			zap.String("observation_health", string(e.latest.Health)),
			zap.Int64("observation_valid_for_seconds", int64(e.latest.validFor()/time.Second)),
		)
	}
	a.l.Info("flow telemetry", fields...)
}
