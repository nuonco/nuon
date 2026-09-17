package telemetry

import (
	"context"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestRuntimeMetrics(t *testing.T) {
	processObserved := time.Now()
	t.Setenv("OTEL_GO_X_DEPRECATED_RUNTIME_METRICS", "false")
	oldGC := debug.SetGCPercent(137)
	oldLimit := debug.SetMemoryLimit(768 << 20)
	oldProcs := runtime.GOMAXPROCS(3)
	t.Cleanup(func() {
		debug.SetGCPercent(oldGC)
		debug.SetMemoryLimit(oldLimit)
		runtime.GOMAXPROCS(oldProcs)
	})
	blocked := make(chan struct{})
	var ready, done sync.WaitGroup
	ready.Add(23)
	done.Add(23)
	for range 23 {
		go func() {
			defer done.Done()
			ready.Done()
			<-blocked
		}()
	}
	t.Cleanup(func() { close(blocked); done.Wait() })
	ready.Wait()
	allocation := make([]byte, 4<<20)
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	globalMeter, globalTracer := otel.GetMeterProvider(), otel.GetTracerProvider()
	time.Sleep(100 * time.Millisecond)
	minimumUptime := time.Since(processObserved).Seconds()
	require.NoError(t, StartRuntimeMetrics(&Config{Endpoint: "http://example.invalid"}, provider))
	require.Same(t, globalMeter, otel.GetMeterProvider())
	require.Same(t, globalTracer, otel.GetTracerProvider())
	collect := func() map[string]metricdata.Metrics {
		var data metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(context.Background(), &data))
		metrics := map[string]metricdata.Metrics{}
		for _, scope := range data.ScopeMetrics {
			for _, m := range scope.Metrics {
				require.NotContains(t, metrics, m.Name)
				metrics[m.Name] = m
			}
		}
		return metrics
	}
	data := collect()
	runtime.KeepAlive(allocation)
	require.Len(t, data, 9)
	for _, spec := range []struct {
		name, unit string
		monotonic  bool
		points     int
	}{
		{"go.memory.used", "By", false, 2},
		{"go.memory.limit", "By", false, 1},
		{"go.memory.allocated", "By", true, 1},
		{"go.memory.allocations", "{allocation}", true, 1},
		{"go.memory.gc.goal", "By", false, 1},
		{"go.goroutine.count", "{goroutine}", false, 1},
		{"go.processor.limit", "{thread}", false, 1},
		{"go.config.gogc", "%", false, 1},
	} {
		require.Contains(t, data, spec.name)
		m := data[spec.name]
		require.Equal(t, spec.unit, m.Unit, spec.name)
		sum := m.Data.(metricdata.Sum[int64])
		require.Equal(t, spec.monotonic, sum.IsMonotonic, spec.name)
		require.Equal(t, metricdata.CumulativeTemporality, sum.Temporality, spec.name)
		require.Len(t, sum.DataPoints, spec.points, spec.name)
		for _, point := range sum.DataPoints {
			require.Positive(t, point.Value, spec.name)
			if spec.name != "go.memory.used" {
				require.Zero(t, point.Attributes.Len(), spec.name)
			}
		}
	}
	value := func(name string) int64 { return data[name].Data.(metricdata.Sum[int64]).DataPoints[0].Value }
	require.EqualValues(t, 137, value("go.config.gogc"))
	require.EqualValues(t, 768<<20, value("go.memory.limit"))
	require.EqualValues(t, 3, value("go.processor.limit"))
	require.GreaterOrEqual(t, value("go.goroutine.count"), int64(23))
	require.GreaterOrEqual(t, value("go.memory.allocated"), int64(len(allocation)))
	memoryTypes := []attribute.Set{}
	for _, point := range data["go.memory.used"].Data.(metricdata.Sum[int64]).DataPoints {
		memoryTypes = append(memoryTypes, point.Attributes)
	}
	require.ElementsMatch(t, []attribute.Set{
		attribute.NewSet(attribute.String("go.memory.type", "stack")),
		attribute.NewSet(attribute.String("go.memory.type", "other")),
	}, memoryTypes)
	uptime := data["process.uptime"].Data.(metricdata.Gauge[float64])
	require.Equal(t, "s", data["process.uptime"].Unit)
	require.Len(t, uptime.DataPoints, 1)
	require.Zero(t, uptime.DataPoints[0].Attributes.Len())
	require.GreaterOrEqual(t, uptime.DataPoints[0].Value, minimumUptime, "uptime includes time before instrumentation initialized")

	debug.SetGCPercent(163)
	time.Sleep(20 * time.Millisecond)
	data = collect()
	require.EqualValues(t, 137, value("go.config.gogc"), "runtime library caches snapshots for 15 seconds")
	require.Greater(t, data["process.uptime"].Data.(metricdata.Gauge[float64]).DataPoints[0].Value, uptime.DataPoints[0].Value)
}

func TestRuntimeMetricsWithoutEndpoint(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	require.NoError(t, StartRuntimeMetrics(&Config{}, provider))
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	require.Empty(t, data.ScopeMetrics)
}
