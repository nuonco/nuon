package health

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	postgresDependency = iota
	clickhouseDependency
	temporalDependency
	dependencyCount
)

var dependencyNames = [dependencyCount]string{"postgresql", "clickhouse", "temporal"}

type dependencyCheck struct {
	started  time.Time
	finished time.Time
	reason   string
}

func (c *dependencyCheck) finish(reason string) {
	c.finished = time.Now()
	c.reason = reason
}

type dependencyState struct {
	completed time.Time
	succeeded time.Time
	healthy   bool
}

type healthMetrics struct {
	checks   metric.Int64Counter
	duration metric.Float64Histogram
	mu       sync.Mutex
	states   [dependencyCount]dependencyState
}

func newHealthMetrics(provider metric.MeterProvider) (*healthMetrics, error) {
	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/health")
	checks, err := meter.Int64Counter("nuon.dependency.checks",
		metric.WithUnit("{check}"),
		metric.WithDescription("Dependency readiness checks by outcome and failure stage; includes skipped checks."))
	if err != nil {
		return nil, err
	}
	duration, err := meter.Float64Histogram("nuon.dependency.check.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of an attempted dependency readiness check."),
		metric.WithExplicitBucketBoundaries(.01, .05, .1, .5, 1, 5))
	if err != nil {
		return nil, err
	}
	status, err := meter.Int64ObservableGauge("nuon.dependency.check.status",
		metric.WithDescription("Last completed dependency readiness check: 1 for success, 0 for failure."))
	if err != nil {
		return nil, err
	}
	completed, err := meter.Float64ObservableGauge("nuon.dependency.check.last_completed",
		metric.WithUnit("s"),
		metric.WithDescription("Unix timestamp of the last completed dependency readiness check."))
	if err != nil {
		return nil, err
	}
	succeeded, err := meter.Float64ObservableGauge("nuon.dependency.check.last_success",
		metric.WithUnit("s"),
		metric.WithDescription("Unix timestamp of the last successful dependency readiness check; absent until first success."))
	if err != nil {
		return nil, err
	}
	m := &healthMetrics{checks: checks, duration: duration}
	_, err = meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		m.mu.Lock()
		states := m.states
		m.mu.Unlock()
		for i, state := range states {
			if state.completed.IsZero() {
				continue
			}
			attrs := metric.WithAttributes(attribute.String("dependency.name", dependencyNames[i]))
			var healthy int64
			if state.healthy {
				healthy = 1
			}
			observer.ObserveInt64(status, healthy, attrs)
			observer.ObserveFloat64(completed, float64(state.completed.UnixNano())/1e9, attrs)
			if !state.succeeded.IsZero() {
				observer.ObserveFloat64(succeeded, float64(state.succeeded.UnixNano())/1e9, attrs)
			}
		}
		return nil
	}, status, completed, succeeded)
	return m, err
}

func (m *healthMetrics) record(ctx context.Context, checks [dependencyCount]dependencyCheck) {
	for i, check := range checks {
		attrs := []attribute.KeyValue{attribute.String("dependency.name", dependencyNames[i])}
		if check.started.IsZero() {
			m.checks.Add(ctx, 1, metric.WithAttributes(append(attrs,
				attribute.String("outcome", "skipped"),
				attribute.String("error.type", "previous_dependency_failed"))...))
			continue
		}
		if check.finished.IsZero() {
			check.finish("incomplete")
		}
		outcome := "success"
		if check.reason != "" {
			outcome = "failure"
		}
		attrs = append(attrs, attribute.String("outcome", outcome))
		m.duration.Record(ctx, check.finished.Sub(check.started).Seconds(), metric.WithAttributes(attrs...))
		if check.reason != "" {
			attrs = append(attrs, attribute.String("error.type", check.reason))
		}
		m.checks.Add(ctx, 1, metric.WithAttributes(attrs...))
		m.mu.Lock()
		state := &m.states[i]
		// Concurrent readiness requests can finish their dependencies in different orders.
		if check.finished.After(state.completed) {
			state.completed = check.finished
			state.healthy = check.reason == ""
		}
		if check.reason == "" && check.finished.After(state.succeeded) {
			state.succeeded = check.finished
		}
		m.mu.Unlock()
	}
}
