package syncer

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	configsync "github.com/nuonco/nuon/pkg/config/sync"
)

func TestMetricsContract(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	metrics := NewMetrics(provider)

	metrics.record(context.Background(), time.Now(), "load", nil)
	metrics.record(context.Background(), time.Now(), "decode", errors.New("invalid config"))
	metrics.record(context.Background(), time.Now(), "sync_transaction", verifiedSyncRejection(configsync.SyncErr{Resource: "component", Description: "unsupported"}))
	metrics.record(context.Background(), time.Now(), "deferred_queues", context.Canceled)

	var collected metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &collected))
	require.Len(t, collected.ScopeMetrics, 1)
	require.Equal(t, "github.com/nuonco/nuon/ctl-api/config-sync", collected.ScopeMetrics[0].Scope.Name)

	byName := make(map[string]metricdata.Metrics)
	for _, metric := range collected.ScopeMetrics[0].Metrics {
		byName[metric.Name] = metric
	}

	attempts := byName["nuon.config.sync.attempts"]
	require.Equal(t, "{attempt}", attempts.Unit)
	points := attempts.Data.(metricdata.Sum[int64]).DataPoints
	require.Len(t, points, 4)
	require.Equal(t, map[attribute.Set]int64{
		attribute.NewSet(attribute.String("outcome", "success"), attribute.String("stage", "none")):              1,
		attribute.NewSet(attribute.String("outcome", "error"), attribute.String("stage", "decode")):              1,
		attribute.NewSet(attribute.String("outcome", "rejected"), attribute.String("stage", "sync_transaction")): 1,
		attribute.NewSet(attribute.String("outcome", "cancelled"), attribute.String("stage", "deferred_queues")): 1,
	}, int64Points(points))

	duration := byName["nuon.config.sync.duration"]
	require.Equal(t, "s", duration.Unit)
	durationPoints := duration.Data.(metricdata.Histogram[float64]).DataPoints
	require.Len(t, durationPoints, 4)
	for _, point := range durationPoints {
		require.Equal(t, []float64{.1, .5, 1, 5, 15, 30, 60, 120, 300, 600}, point.Bounds)
		require.Equal(t, 1, point.Attributes.Len(), "duration must only carry outcome")
		require.EqualValues(t, 1, point.Count)
	}
}

func TestMetricsClassificationIsConservative(t *testing.T) {
	rejection := configsync.SyncErr{Resource: "branch", Description: "invalid"}
	wrappedRejection := verifiedSyncRejection(fmt.Errorf("validation failed: %w", rejection))
	var preserved configsync.SyncErr
	require.ErrorAs(t, wrappedRejection, &preserved)

	tests := []struct {
		name string
		ctx  context.Context
		err  error
		want string
	}{
		{name: "verified rejection", ctx: context.Background(), err: wrappedRejection, want: "rejected"},
		{name: "legacy sync error remains error", ctx: context.Background(), err: rejection, want: "error"},
		{name: "wrapped legacy sync error remains error", ctx: context.Background(), err: fmt.Errorf("lookup failed: %w", rejection), want: "error"},
		{name: "cancelled error", ctx: context.Background(), err: fmt.Errorf("run stopped: %w", context.Canceled), want: "cancelled"},
		{name: "cancelled context wins", ctx: cancelledContext(), err: wrappedRejection, want: "cancelled"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, (*Metrics)(nil).record(test.ctx, time.Now(), "sync_transaction", test.err))
		})
	}
}

func int64Points(points []metricdata.DataPoint[int64]) map[attribute.Set]int64 {
	result := make(map[attribute.Set]int64, len(points))
	for _, point := range points {
		result[point.Attributes] = point.Value
	}
	return result
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
