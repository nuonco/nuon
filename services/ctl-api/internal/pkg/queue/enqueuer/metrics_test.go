package enqueuer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
)

func TestEnqueuerMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })

	backlog := 3
	metrics := newEnqueuerMetrics(provider, func() int { return backlog })
	require.NotNil(t, metrics)
	metrics.recordDispatch(context.Background(), EnqueueSourceAwait, time.Now(), nil)
	metrics.recordDispatch(context.Background(), "unbounded-value", time.Now(), errors.New("dispatch failed"))
	metrics.recordOperation(context.Background(), EnqueueSourceSweep, enqueueOperationMarkEnqueued, errors.New("database failed"))
	metrics.recordOperation(context.Background(), EnqueueSourceSweep, enqueueOperationUpdateMetadata, nil)
	metrics.workerStarted()
	metrics.channelDropped(context.Background())

	data := collectEnqueuerMetrics(t, reader)
	require.Equal(t, int64(3), gaugeValue(t, data, "nuon.queue.enqueuer.local.backlog"))
	require.Equal(t, int64(1), gaugeValue(t, data, "nuon.queue.enqueuer.local.workers.active"))
	require.Equal(t, int64(1), sumTotal(t, data, "nuon.queue.enqueuer.channel.dropped"))
	require.Equal(t, int64(2), sumTotal(t, data, "nuon.queue.enqueuer.dispatch.attempts"))
	require.Equal(t, uint64(2), histogramTotal(t, data, "nuon.queue.enqueuer.dispatch.duration"))
	require.Equal(t, int64(2), sumTotal(t, data, "nuon.queue.enqueuer.operations"))

	dispatch := findMetric(t, data, "nuon.queue.enqueuer.dispatch.attempts").Data.(metricdata.Sum[int64])
	want := map[string]int64{
		attributeSetKey(attribute.NewSet(
			attribute.String("nuon.queue.enqueuer.source", EnqueueSourceAwait),
			attribute.String("nuon.queue.enqueuer.outcome", "success"),
		)): 1,
		attributeSetKey(attribute.NewSet(
			attribute.String("nuon.queue.enqueuer.source", enqueueSourceOther),
			attribute.String("nuon.queue.enqueuer.outcome", "failure"),
		)): 1,
	}
	for _, point := range dispatch.DataPoints {
		require.Equal(t, want[point.Attributes.Encoded(attribute.DefaultEncoder())], point.Value)
		require.Equal(t, 2, point.Attributes.Len())
	}
}

func TestSendRecordsDroppedLocalChannelItem(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })

	e := &Enqueuer{ch: make(chan string, 1), l: zap.NewNop()}
	e.metrics = newEnqueuerMetrics(provider, func() int { return len(e.ch) })
	e.Send("first")
	e.Send("dropped")

	data := collectEnqueuerMetrics(t, reader)
	require.Equal(t, int64(1), gaugeValue(t, data, "nuon.queue.enqueuer.local.backlog"))
	require.Equal(t, int64(1), sumTotal(t, data, "nuon.queue.enqueuer.channel.dropped"))
}

func TestEnqueuerMetricsNilProvider(t *testing.T) {
	require.NotPanics(t, func() {
		var metrics *enqueuerMetrics
		metrics.recordDispatch(context.Background(), EnqueueSourceChannel, time.Now(), nil)
		metrics.recordOperation(context.Background(), EnqueueSourceChannel, enqueueOperationMarkEnqueued, nil)
		metrics.workerStarted()
		metrics.workerFinished()
		metrics.channelDropped(context.Background())
		metrics.close()
	})
}

func collectEnqueuerMetrics(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	return data
}

func findMetric(t *testing.T, data metricdata.ResourceMetrics, name string) metricdata.Metrics {
	t.Helper()
	for _, scope := range data.ScopeMetrics {
		for _, collected := range scope.Metrics {
			if collected.Name == name {
				return collected
			}
		}
	}
	t.Fatalf("metric %q not found", name)
	return metricdata.Metrics{}
}

func sumTotal(t *testing.T, data metricdata.ResourceMetrics, name string) int64 {
	t.Helper()
	var total int64
	for _, point := range findMetric(t, data, name).Data.(metricdata.Sum[int64]).DataPoints {
		total += point.Value
	}
	return total
}

func histogramTotal(t *testing.T, data metricdata.ResourceMetrics, name string) uint64 {
	t.Helper()
	var total uint64
	for _, point := range findMetric(t, data, name).Data.(metricdata.Histogram[float64]).DataPoints {
		total += point.Count
	}
	return total
}

func gaugeValue(t *testing.T, data metricdata.ResourceMetrics, name string) int64 {
	t.Helper()
	points := findMetric(t, data, name).Data.(metricdata.Gauge[int64]).DataPoints
	require.Len(t, points, 1)
	return points[0].Value
}

func attributeSetKey(set attribute.Set) string {
	return set.Encoded(attribute.DefaultEncoder())
}
