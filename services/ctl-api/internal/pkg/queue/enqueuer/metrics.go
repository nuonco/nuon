package enqueuer

import (
	"context"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	enqueueSourceOther = "other"

	enqueueOperationMarkEnqueued   = "mark_enqueued"
	enqueueOperationUpdateMetadata = "update_metadata"
	enqueueOperationOther          = "other"
)

type enqueuerMetrics struct {
	dispatchAttempts metric.Int64Counter
	dispatchDuration metric.Float64Histogram
	operations       metric.Int64Counter
	dropped          metric.Int64Counter
	backlog          metric.Int64ObservableGauge
	activeWorkers    metric.Int64ObservableGauge
	registration     metric.Registration
	active           atomic.Int64
}

func newEnqueuerMetrics(provider metric.MeterProvider, backlog func() int) *enqueuerMetrics {
	if provider == nil {
		return nil
	}

	meter := provider.Meter("github.com/nuonco/nuon/services/ctl-api/queue-enqueuer")
	dispatchAttempts, err := meter.Int64Counter("nuon.queue.enqueuer.dispatch.attempts", metric.WithUnit("{attempt}"))
	if err != nil {
		return nil
	}
	dispatchDuration, err := meter.Float64Histogram(
		"nuon.queue.enqueuer.dispatch.duration",
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30),
	)
	if err != nil {
		return nil
	}
	operations, err := meter.Int64Counter("nuon.queue.enqueuer.operations", metric.WithUnit("{operation}"))
	if err != nil {
		return nil
	}
	dropped, err := meter.Int64Counter("nuon.queue.enqueuer.channel.dropped", metric.WithUnit("{signal}"))
	if err != nil {
		return nil
	}
	backlogGauge, err := meter.Int64ObservableGauge("nuon.queue.enqueuer.local.backlog", metric.WithUnit("{signal}"))
	if err != nil {
		return nil
	}
	activeWorkers, err := meter.Int64ObservableGauge("nuon.queue.enqueuer.local.workers.active", metric.WithUnit("{worker}"))
	if err != nil {
		return nil
	}

	m := &enqueuerMetrics{
		dispatchAttempts: dispatchAttempts,
		dispatchDuration: dispatchDuration,
		operations:       operations,
		dropped:          dropped,
		backlog:          backlogGauge,
		activeWorkers:    activeWorkers,
	}
	m.registration, err = meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		observer.ObserveInt64(m.backlog, int64(backlog()))
		observer.ObserveInt64(m.activeWorkers, m.active.Load())
		return nil
	}, m.backlog, m.activeWorkers)
	if err != nil {
		return nil
	}
	return m
}

func boundedEnqueueSource(source string) string {
	switch source {
	case EnqueueSourceChannel, EnqueueSourceAwait, EnqueueSourceSweep:
		return source
	default:
		return enqueueSourceOther
	}
}

func boundedEnqueueOperation(operation string) string {
	switch operation {
	case enqueueOperationMarkEnqueued, enqueueOperationUpdateMetadata:
		return operation
	default:
		return enqueueOperationOther
	}
}

func (m *enqueuerMetrics) recordDispatch(ctx context.Context, source string, started time.Time, err error) {
	if m == nil {
		return
	}
	outcome := "success"
	if err != nil {
		outcome = "failure"
	}
	attrs := metric.WithAttributes(
		attribute.String("nuon.queue.enqueuer.source", boundedEnqueueSource(source)),
		attribute.String("nuon.queue.enqueuer.outcome", outcome),
	)
	m.dispatchAttempts.Add(ctx, 1, attrs)
	m.dispatchDuration.Record(ctx, time.Since(started).Seconds(), attrs)
}

func (m *enqueuerMetrics) recordOperation(ctx context.Context, source, operation string, err error) {
	if m == nil {
		return
	}
	outcome := "success"
	if err != nil {
		outcome = "failure"
	}
	m.operations.Add(ctx, 1, metric.WithAttributes(
		attribute.String("nuon.queue.enqueuer.source", boundedEnqueueSource(source)),
		attribute.String("nuon.queue.enqueuer.operation", boundedEnqueueOperation(operation)),
		attribute.String("nuon.queue.enqueuer.outcome", outcome),
	))
}

func (m *enqueuerMetrics) workerStarted() {
	if m != nil {
		m.active.Add(1)
	}
}

func (m *enqueuerMetrics) workerFinished() {
	if m != nil {
		m.active.Add(-1)
	}
}

func (m *enqueuerMetrics) channelDropped(ctx context.Context) {
	if m != nil {
		m.dropped.Add(ctx, 1)
	}
}

func (m *enqueuerMetrics) close() {
	if m != nil && m.registration != nil {
		_ = m.registration.Unregister()
	}
}
