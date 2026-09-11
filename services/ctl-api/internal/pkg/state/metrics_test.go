package state

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestOperationMetrics(t *testing.T) {
	ctx := context.Background()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(ctx)) })
	metrics := NewMetrics(provider)
	metrics.Record(ctx, "get", time.Now().Add(-2*time.Second), nil)
	metrics.Record(ctx, "get", time.Now(), nil)
	metrics.Record(ctx, "save", time.Now(), errors.New("private error"))
	NewMetrics(nil).Record(ctx, "save", time.Now(), nil)

	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &data))
	require.Len(t, data.ScopeMetrics, 1)
	require.Len(t, data.ScopeMetrics[0].Metrics, 2)
	for _, m := range data.ScopeMetrics[0].Metrics {
		switch m.Name {
		case "nuon.install.state.operations":
			counts := map[string]int64{}
			for _, point := range m.Data.(metricdata.Sum[int64]).DataPoints {
				require.Equal(t, 2, point.Attributes.Len())
				op, _ := point.Attributes.Value("operation")
				outcome, _ := point.Attributes.Value("outcome")
				counts[op.AsString()+":"+outcome.AsString()] = point.Value
			}
			require.Equal(t, map[string]int64{"get:success": 2, "save:error": 1}, counts)
		case "nuon.install.state.operation.duration":
			require.Equal(t, "s", m.Unit)
			for _, point := range m.Data.(metricdata.Histogram[float64]).DataPoints {
				op, _ := point.Attributes.Value("operation")
				if op.AsString() == "get" {
					require.EqualValues(t, 2, point.Count)
					require.GreaterOrEqual(t, point.Sum, float64(2))
					require.Less(t, point.Sum, float64(20))
				}
			}
		default:
			t.Fatalf("unexpected metric %q", m.Name)
		}
	}
}
