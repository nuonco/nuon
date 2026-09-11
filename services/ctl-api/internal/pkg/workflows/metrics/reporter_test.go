package workflowmetrics

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
)

func testReporter(t *testing.T) (*reporter, *sdkmetric.ManualReader) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	r, err := newReporter(provider, zap.NewNop())
	require.NoError(t, err)
	return r, reader
}

func collect(t *testing.T, reader *sdkmetric.ManualReader) map[string]metricdata.Metrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))
	result := make(map[string]metricdata.Metrics)
	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			result[m.Name] = m
		}
	}
	return result
}

func TestSnapshotObservation(t *testing.T) {
	r, reader := testReporter(t)
	require.Empty(t, collect(t, reader), "never collected is not a healthy zero")
	now := time.Now()
	oldest := now.Add(-37 * time.Minute).Truncate(time.Second)
	r.values = map[bucket]value{{"manual_deploy", "awaiting_retry"}: {3, oldest}}
	r.collectedAt, r.active = now, true

	metrics := collect(t, reader)
	counts := metrics["nuon.workflow.current"].Data.(metricdata.Gauge[int64]).DataPoints
	require.Len(t, counts, 12)
	for _, p := range counts {
		typ, _ := p.Attributes.Value("workflow.type")
		state, _ := p.Attributes.Value("workflow.state")
		require.Equal(t, 2, p.Attributes.Len())
		if typ.AsString() == "manual_deploy" && state.AsString() == "awaiting_retry" {
			require.EqualValues(t, 3, p.Value)
		} else {
			require.Zero(t, p.Value)
		}
	}
	ages := metrics["nuon.workflow.oldest_created_at"].Data.(metricdata.Gauge[float64]).DataPoints
	require.Len(t, ages, 1)
	require.Equal(t, float64(oldest.Unix()), ages[0].Value)

	r.values = nil
	metrics = collect(t, reader)
	for _, p := range metrics["nuon.workflow.current"].Data.(metricdata.Gauge[int64]).DataPoints {
		require.Zero(t, p.Value, "last workflow leaving a category clears its count")
	}
	require.NotContains(t, metrics, "nuon.workflow.oldest_created_at")

	r.collectedAt = now.Add(-170 * time.Second)
	require.Contains(t, collect(t, reader), "nuon.workflow.current", "freshness allows a 60-second export interval plus query and delivery delay")

	r.collectedAt = now.Add(-snapshotTTL - time.Second)
	metrics = collect(t, reader)
	require.Len(t, metrics, 1, "expired snapshots expose only actual collection time")
	stamp := metrics["nuon.workflow.snapshot.collected_at"].Data.(metricdata.Gauge[float64]).DataPoints[0].Value
	require.Equal(t, float64(r.collectedAt.UnixNano())/1e9, stamp)

	r.collectedAt = now
	r.release(nil)
	require.Len(t, collect(t, reader), 1, "former leaders must stop observing counts immediately")
}

func TestDisabledDoesNotStartReporter(t *testing.T) {
	// Nil dependencies prove disabled export neither registers a callback nor opens a connection.
	require.NoError(t, Start(nil, nil, &telemetry.Config{}, nil, nil))
}

func TestCancellationDuringConnection(t *testing.T) {
	r, reader := testReporter(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	connecting, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		r.run(ctx, func(ctx context.Context) (*pgx.Conn, error) {
			close(connecting)
			<-ctx.Done()
			return nil, ctx.Err()
		})
	}()
	select {
	case <-connecting:
	case <-time.After(time.Second):
		t.Fatal("reporter did not start connecting")
	}
	require.Empty(t, collect(t, reader), "export must not wait for the database")
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("reporter did not stop after cancellation")
	}
}
