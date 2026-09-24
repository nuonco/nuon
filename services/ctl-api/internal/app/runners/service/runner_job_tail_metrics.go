package service

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type runnerJobTailMetrics struct {
	sessions metric.Int64Counter
	probes   metric.Int64Counter
	wakes    metric.Int64Counter
}

func newRunnerJobTailMetrics(provider metric.MeterProvider) *runnerJobTailMetrics {
	if provider == nil {
		return nil
	}
	meter := provider.Meter("github.com/nuonco/nuon/ctl-api/runner-job-tail")
	sessions, _ := meter.Int64Counter("nuon.runner.job_tail.sessions", metric.WithUnit("{session}"), metric.WithDescription("Completed long-poll sessions after request validation."))
	probes, _ := meter.Int64Counter("nuon.runner.job_tail.probes", metric.WithUnit("{probe}"), metric.WithDescription("Long-poll probe attempts, including semaphore cancellation and retries."))
	wakes, _ := meter.Int64Counter("nuon.runner.job_tail.notification.wakes", metric.WithUnit("{wake}"), metric.WithDescription("Notifications consumed by parked long-poll handlers."))
	m := &runnerJobTailMetrics{sessions: sessions, probes: probes, wakes: wakes}
	for _, outcome := range []string{jobTailOutcomeHotHit, jobTailOutcomeIdleThenHit, jobTailOutcomeTimeoutEmpty, jobTailOutcomeClientCancel, jobTailOutcomeError} {
		sessions.Add(context.Background(), 0, metric.WithAttributes(attribute.String("outcome", outcome)))
	}
	for _, outcome := range []string{"hit", "empty", "transient_error", "error", "cancelled"} {
		probes.Add(context.Background(), 0, metric.WithAttributes(attribute.String("outcome", outcome)))
	}
	wakes.Add(context.Background(), 0)
	return m
}

func (m *runnerJobTailMetrics) session(outcome string) {
	if m != nil {
		m.sessions.Add(context.Background(), 1, metric.WithAttributes(attribute.String("outcome", outcome)))
	}
}

func (m *runnerJobTailMetrics) probe(ctx context.Context, job *app.RunnerJob, err error) {
	if m == nil {
		return
	}
	outcome := "empty"
	if err != nil {
		switch {
		case ctx.Err() != nil:
			outcome = "cancelled"
		case isTransientTailProbeError(err):
			outcome = "transient_error"
		default:
			outcome = "error"
		}
	} else if job != nil {
		outcome = "hit"
	}
	m.probes.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
}

func (m *runnerJobTailMetrics) wake(ctx context.Context) {
	if m != nil {
		m.wakes.Add(ctx, 1)
	}
}

type runnerJobListenerMetrics struct {
	connected     atomic.Int64
	failures      metric.Int64Counter
	notifications metric.Int64Counter
}

func newRunnerJobListenerMetrics(provider metric.MeterProvider) *runnerJobListenerMetrics {
	if provider == nil {
		return nil
	}
	m := &runnerJobListenerMetrics{}
	meter := provider.Meter("github.com/nuonco/nuon/ctl-api/runner-job-tail")
	_, _ = meter.Int64ObservableGauge("nuon.runner.job_tail.listener.connected", metric.WithUnit("1"),
		metric.WithDescription("Whether this process has an active LISTEN subscription; idle connections remain observable."),
		metric.WithInt64Callback(func(_ context.Context, observer metric.Int64Observer) error {
			observer.Observe(m.connected.Load())
			return nil
		}))
	m.failures, _ = meter.Int64Counter("nuon.runner.job_tail.listener.failures", metric.WithUnit("{failure}"), metric.WithDescription("Failed listener sessions, including connect failures, excluding rotation and shutdown."))
	m.notifications, _ = meter.Int64Counter("nuon.runner.job_tail.listener.notifications", metric.WithUnit("{notification}"), metric.WithDescription("Received notifications by payload validity; not unique jobs or handler wakes."))
	m.failures.Add(context.Background(), 0)
	for _, outcome := range []string{"valid", "invalid"} {
		m.notifications.Add(context.Background(), 0, metric.WithAttributes(attribute.String("outcome", outcome)))
	}
	return m
}

func (m *runnerJobListenerMetrics) setConnected(value int64) {
	if m != nil {
		m.connected.Store(value)
	}
}

func (m *runnerJobListenerMetrics) failure(ctx context.Context) {
	if m != nil {
		m.failures.Add(ctx, 1)
	}
}

func (m *runnerJobListenerMetrics) notification(ctx context.Context, outcome string) {
	if m != nil {
		m.notifications.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
	}
}
