package health

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func collectHealthMetrics(t *testing.T, reader *sdkmetric.ManualReader) map[string]metricdata.Metrics {
	t.Helper()
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	result := map[string]metricdata.Metrics{}
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			result[m.Name] = m
		}
	}
	return result
}

func TestHealthMetricsFreshness(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	m, err := newHealthMetrics(provider)
	require.NoError(t, err)
	require.Empty(t, collectHealthMetrics(t, reader), "no check must not look healthy or fresh")
	ctx := context.Background()
	m.record(ctx, [dependencyCount]dependencyCheck{
		{started: time.Unix(100, 0), finished: time.Unix(102, 0), reason: "ping"},
	})
	data := collectHealthMetrics(t, reader)
	require.NotContains(t, data, "nuon.dependency.check.last_success")
	status := data["nuon.dependency.check.status"].Data.(metricdata.Gauge[int64]).DataPoints
	require.Len(t, status, 1)
	require.Zero(t, status[0].Value)
	require.Equal(t, attribute.NewSet(attribute.String("dependency.name", "postgresql")), status[0].Attributes)

	m.record(ctx, [dependencyCount]dependencyCheck{
		{started: time.Unix(109, 0), finished: time.Unix(110, 0)},
		{started: time.Unix(110, 0), finished: time.Unix(113, 0)},
		{started: time.Unix(113, 0), finished: time.Unix(117, 0)},
	})
	m.record(ctx, [dependencyCount]dependencyCheck{
		{started: time.Unix(118, 0), finished: time.Unix(120, 0), reason: "connection"},
	})
	data = collectHealthMetrics(t, reader)
	lastSuccess := map[string]float64{}
	for _, p := range data["nuon.dependency.check.last_success"].Data.(metricdata.Gauge[float64]).DataPoints {
		dep, _ := p.Attributes.Value("dependency.name")
		lastSuccess[dep.AsString()] = p.Value
	}
	require.Equal(t, map[string]float64{"postgresql": 110, "clickhouse": 113, "temporal": 117}, lastSuccess)
	// An earlier request can publish after a later one without making state regress.
	m.record(ctx, [dependencyCount]dependencyCheck{
		{started: time.Unix(111, 0), finished: time.Unix(114, 0)},
		{started: time.Unix(114, 0), finished: time.Unix(115, 0)},
		{started: time.Unix(115, 0), finished: time.Unix(116, 0)},
	})
	for range 2 {
		data = collectHealthMetrics(t, reader)
		require.Len(t, data, 5)
		for name, want := range map[string]map[string]float64{
			"nuon.dependency.check.last_completed": {"postgresql": 120, "clickhouse": 115, "temporal": 117},
			"nuon.dependency.check.last_success":   {"postgresql": 114, "clickhouse": 115, "temporal": 117},
		} {
			got := map[string]float64{}
			for _, p := range data[name].Data.(metricdata.Gauge[float64]).DataPoints {
				require.Equal(t, 1, p.Attributes.Len())
				dep, _ := p.Attributes.Value("dependency.name")
				got[dep.AsString()] = p.Value
			}
			require.Equal(t, want, got)
			require.Equal(t, "s", data[name].Unit)
		}
		states := map[string]int64{}
		for _, p := range data["nuon.dependency.check.status"].Data.(metricdata.Gauge[int64]).DataPoints {
			dep, _ := p.Attributes.Value("dependency.name")
			states[dep.AsString()] = p.Value
		}
		require.Equal(t, map[string]int64{"postgresql": 0, "clickhouse": 1, "temporal": 1}, states)
		sum := data["nuon.dependency.checks"].Data.(metricdata.Sum[int64])
		require.True(t, sum.IsMonotonic)
		require.Equal(t, metricdata.CumulativeTemporality, sum.Temporality)
		counts := map[string]int64{}
		for _, p := range sum.DataPoints {
			dep, _ := p.Attributes.Value("dependency.name")
			outcome, _ := p.Attributes.Value("outcome")
			counts[dep.AsString()+"/"+outcome.AsString()] += p.Value
		}
		require.Equal(t, map[string]int64{
			"postgresql/success": 2, "postgresql/failure": 2,
			"clickhouse/success": 2, "clickhouse/skipped": 2,
			"temporal/success": 2, "temporal/skipped": 2,
		}, counts)
		durations := map[string]float64{}
		for _, p := range data["nuon.dependency.check.duration"].Data.(metricdata.Histogram[float64]).DataPoints {
			dep, _ := p.Attributes.Value("dependency.name")
			outcome, _ := p.Attributes.Value("outcome")
			require.NotEqual(t, "skipped", outcome.AsString())
			require.Equal(t, 2, p.Attributes.Len())
			require.EqualValues(t, 2, p.Count)
			require.Equal(t, []float64{.01, .05, .1, .5, 1, 5}, p.Bounds)
			durations[dep.AsString()+"/"+outcome.AsString()] = p.Sum
		}
		require.Equal(t, map[string]float64{
			"postgresql/success": 4, "postgresql/failure": 4,
			"clickhouse/success": 4, "temporal/success": 5,
		}, durations)
	}
}

func TestHealthMetricsConcurrentCollection(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	m, err := newHealthMetrics(provider)
	require.NoError(t, err)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			var checks [dependencyCount]dependencyCheck
			for d := range checks {
				checks[d] = dependencyCheck{started: time.Unix(int64(i), 0), finished: time.Unix(int64(i+1), 0)}
			}
			m.record(context.Background(), checks)
		})
	}
	for range 5 {
		collectHealthMetrics(t, reader)
	}
	wg.Wait()
	data := collectHealthMetrics(t, reader)
	points := data["nuon.dependency.checks"].Data.(metricdata.Sum[int64]).DataPoints
	require.Len(t, points, 3)
	for _, p := range points {
		require.EqualValues(t, 20, p.Value)
	}
	for _, p := range data["nuon.dependency.check.last_completed"].Data.(metricdata.Gauge[float64]).DataPoints {
		require.Equal(t, float64(20), p.Value)
	}
}

func TestHealthMetricsNoop(t *testing.T) {
	m, err := newHealthMetrics(noop.NewMeterProvider())
	require.NoError(t, err)
	m.record(context.Background(), [dependencyCount]dependencyCheck{
		{started: time.Now().Add(-time.Second), finished: time.Now()},
	})
}
