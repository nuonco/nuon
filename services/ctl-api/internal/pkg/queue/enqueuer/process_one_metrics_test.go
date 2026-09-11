package enqueuer

import (
	"context"
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	basemetrics "github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type enqueueInlineTemporalClient struct {
	temporalclient.Client
	calls int
	err   error
}

func (c *enqueueInlineTemporalClient) SignalWithStartWorkflowInNamespace(context.Context, string, string, string, interface{}, tclient.StartWorkflowOptions, interface{}, interface{}) (tclient.WorkflowRun, error) {
	c.calls++
	return nil, c.err
}

func TestEnqueueInlineMetricsCallPaths(t *testing.T) {
	t.Run("already enqueued skips RPC and metrics", func(t *testing.T) {
		e, client, reader, shutdown := newEnqueueInlineMetricsHarness(t, true, nil, nil, nil)
		defer shutdown()

		err := e.EnqueueInline(context.Background(), "signal-id", EnqueueSourceAwait)
		require.NoError(t, err)
		require.Zero(t, client.calls)
		require.Empty(t, enqueueInlineMetricPoints(t, reader, "nuon.queue.enqueuer.dispatch.attempts"))
		require.Empty(t, enqueueInlineMetricPoints(t, reader, "nuon.queue.enqueuer.operations"))
	})

	t.Run("Temporal failure records dispatch and preserves error", func(t *testing.T) {
		wantErr := errors.New("Temporal unavailable")
		e, client, reader, shutdown := newEnqueueInlineMetricsHarness(t, false, wantErr, nil, nil)
		defer shutdown()

		err := e.EnqueueInline(context.Background(), "signal-id", EnqueueSourceAwait)
		require.ErrorIs(t, err, wantErr)
		require.Equal(t, 1, client.calls)
		require.Equal(t, map[string]int64{"await/failure": 1}, enqueueInlineDispatchCounts(t, reader))
		require.Equal(t, map[string]int64{"await/update_metadata/success": 1}, enqueueInlineOperationCounts(t, reader))
	})

	t.Run("mark enqueued failure records operation and preserves success", func(t *testing.T) {
		wantErr := errors.New("mark enqueued failed")
		e, client, reader, shutdown := newEnqueueInlineMetricsHarness(t, false, nil, wantErr, nil)
		defer shutdown()

		err := e.EnqueueInline(context.Background(), "signal-id", EnqueueSourceSweep)
		require.NoError(t, err)
		require.Equal(t, 1, client.calls)
		require.Equal(t, map[string]int64{"sweep/success": 1}, enqueueInlineDispatchCounts(t, reader))
		require.Equal(t, map[string]int64{
			"sweep/mark_enqueued/failure":   1,
			"sweep/update_metadata/success": 1,
		}, enqueueInlineOperationCounts(t, reader))
	})

	t.Run("metadata failure records operation and preserves success", func(t *testing.T) {
		wantErr := errors.New("metadata failed")
		e, client, reader, shutdown := newEnqueueInlineMetricsHarness(t, false, nil, nil, wantErr)
		defer shutdown()

		err := e.EnqueueInline(context.Background(), "signal-id", EnqueueSourceChannel)
		require.NoError(t, err)
		require.Equal(t, 1, client.calls)
		require.Equal(t, map[string]int64{"channel/success": 1}, enqueueInlineDispatchCounts(t, reader))
		require.Equal(t, map[string]int64{
			"channel/mark_enqueued/success":   1,
			"channel/update_metadata/failure": 1,
		}, enqueueInlineOperationCounts(t, reader))
	})
}

func newEnqueueInlineMetricsHarness(t *testing.T, enqueued bool, temporalErr, markErr, metadataErr error) (*Enqueuer, *enqueueInlineTemporalClient, *sdkmetric.ManualReader, func()) {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused"}), &gorm.Config{
		DisableAutomaticPing:   true,
		DryRun:                 true,
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)
	require.NoError(t, db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
		switch dest := tx.Statement.Dest.(type) {
		case *app.QueueSignal:
			dest.ID = "signal-id"
			dest.QueueID = "queue-id"
			dest.Enqueued = enqueued
			dest.Workflow = signaldb.WorkflowRef{ID: "signal-workflow-id"}
		case *app.Queue:
			dest.ID = "queue-id"
			dest.Name = "default"
			dest.Workflow = signaldb.WorkflowRef{ID: "queue-workflow-id", Namespace: "queue-namespace"}
		default:
			t.Fatalf("unexpected query destination %T", dest)
		}
		tx.RowsAffected = 1
	}))
	require.NoError(t, db.Callback().Update().Replace("gorm:update", func(tx *gorm.DB) {
		values, ok := tx.Statement.Dest.(map[string]interface{})
		require.True(t, ok, "unexpected update destination %T", tx.Statement.Dest)
		if _, ok := values["enqueued"]; ok {
			tx.AddError(markErr)
			return
		}
		if _, ok := values["status"]; ok {
			tx.AddError(metadataErr)
			return
		}
		t.Fatalf("unexpected update values %#v", values)
	}))

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	mw, err := basemetrics.New(validator.New(), basemetrics.WithDisable(true), basemetrics.WithLogger(zap.NewNop()))
	require.NoError(t, err)
	client := &enqueueInlineTemporalClient{err: temporalErr}
	e := &Enqueuer{
		db:      db,
		cfg:     &internal.Config{Version: "test"},
		tClient: client,
		l:       zap.NewNop(),
		mw:      mw,
		metrics: newEnqueuerMetrics(provider, func() int { return 0 }),
	}
	return e, client, reader, func() { require.NoError(t, provider.Shutdown(context.Background())) }
}

func enqueueInlineDispatchCounts(t *testing.T, reader *sdkmetric.ManualReader) map[string]int64 {
	t.Helper()
	return enqueueInlineMetricCounts(t, reader, "nuon.queue.enqueuer.dispatch.attempts", []string{
		"nuon.queue.enqueuer.source",
		"nuon.queue.enqueuer.outcome",
	})
}

func enqueueInlineOperationCounts(t *testing.T, reader *sdkmetric.ManualReader) map[string]int64 {
	t.Helper()
	return enqueueInlineMetricCounts(t, reader, "nuon.queue.enqueuer.operations", []string{
		"nuon.queue.enqueuer.source",
		"nuon.queue.enqueuer.operation",
		"nuon.queue.enqueuer.outcome",
	})
}

func enqueueInlineMetricCounts(t *testing.T, reader *sdkmetric.ManualReader, name string, attributes []string) map[string]int64 {
	t.Helper()
	counts := make(map[string]int64)
	for _, point := range enqueueInlineMetricPoints(t, reader, name) {
		parts := make([]string, 0, len(attributes))
		for _, name := range attributes {
			value, ok := point.Attributes.Value(attribute.Key(name))
			require.True(t, ok)
			parts = append(parts, value.AsString())
		}
		key := parts[0]
		for _, part := range parts[1:] {
			key += "/" + part
		}
		counts[key] += point.Value
	}
	return counts
}

func enqueueInlineMetricPoints(t *testing.T, reader *sdkmetric.ManualReader, name string) []metricdata.DataPoint[int64] {
	t.Helper()
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	for _, scope := range data.ScopeMetrics {
		for _, collected := range scope.Metrics {
			if collected.Name == name {
				return collected.Data.(metricdata.Sum[int64]).DataPoints
			}
		}
	}
	return nil
}
