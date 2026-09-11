package poolmetrics

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/fx/fxtest"
)

func postgresPool(t *testing.T, max int32) *pgxpool.Pool {
	t.Helper()
	cfg, err := pgxpool.ParseConfig("postgres://test:password@example.invalid:5432/metrics_fixture")
	require.NoError(t, err)
	cfg.MaxConns = max
	cfg.MinConns = 0
	cfg.MinIdleConns = 0
	cfg.BeforeConnect = func(context.Context, *pgx.ConnConfig) error {
		t.Error("pool observation must not connect to PostgreSQL")
		return errors.New("unexpected database connection")
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

type poolConnector struct{}
type poolDriver struct{}
type poolConn struct{}

func (poolConnector) Connect(context.Context) (driver.Conn, error) { return poolConn{}, nil }
func (poolConnector) Driver() driver.Driver                        { return poolDriver{} }
func (poolDriver) Open(string) (driver.Conn, error)                { return poolConn{}, nil }
func (poolConn) Close() error                                      { return nil }
func (poolConn) Prepare(string) (driver.Stmt, error)               { return nil, errors.New("unexpected SQL") }
func (poolConn) Begin() (driver.Tx, error)                         { return nil, errors.New("unexpected transaction") }

func cancelAcquires(t *testing.T, pool *pgxpool.Pool, count int) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for range count {
		_, err := pool.Acquire(ctx)
		require.ErrorIs(t, err, context.Canceled)
	}
}

func TestPoolOTLPExport(t *testing.T) {
	requests := make(chan []byte, 4)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/metrics" {
			http.NotFound(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		requests <- body
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	t.Cleanup(receiver.Close)
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "")
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "3600000")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "delta")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "nuon.control_plane.id=cp-test,service.instance.id=api-test")
	t.Setenv("OTEL_SERVICE_NAME", "ctl-api-test")
	cfg, err := telemetry.NewConfig(&internal.Config{OTELExporterOTLPEndpoint: receiver.URL})
	require.NoError(t, err)
	lc := fxtest.NewLifecycle(t)
	provider, err := telemetry.NewMeterProvider(lc, cfg)
	require.NoError(t, err)
	m := New(provider)
	globalMeter, globalTracer := otel.GetMeterProvider(), otel.GetTracerProvider()
	roles := []string{"primary", "replica", "admin_replica"}
	pools := make([]*pgxpool.Pool, 0, len(roles))
	for i, role := range roles {
		pool := postgresPool(t, int32(17+i))
		cancelAcquires(t, pool, 10+i)
		require.NoError(t, m.RegisterPostgres(role, pool))
		pools = append(pools, pool)
	}
	db := sql.OpenDB(poolConnector{})
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.SetMaxOpenConns(3)
	for range 2 {
		conn, err := db.Conn(context.Background())
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, conn.Close()) })
	}
	third, err := db.Conn(context.Background())
	require.NoError(t, err)
	waitCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err = db.Conn(waitCtx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.NoError(t, third.Close())
	waitDuration := db.Stats().WaitDuration
	require.Positive(t, waitDuration)
	require.NoError(t, m.RegisterClickHouse(lc, db))
	require.Same(t, globalMeter, otel.GetMeterProvider())
	require.Same(t, globalTracer, otel.GetTracerProvider())
	require.IsType(t, poolDriver{}, db.Driver())
	lc.RequireStart()
	t.Cleanup(func() { lc.RequireStop() })

	for iteration, count := range []int64{10, 10, 13} {
		if iteration == 2 {
			cancelAcquires(t, pools[0], 3)
			time.Sleep(1100 * time.Millisecond) // otelpgx caches snapshots for one second.
		}
		require.NoError(t, provider.(*sdkmetric.MeterProvider).ForceFlush(context.Background()))
		var body []byte
		select {
		case body = <-requests:
		case <-time.After(time.Second):
			t.Fatal("no pool metric export received")
		}
		payload := pmetricotlp.NewExportRequest()
		require.NoError(t, payload.UnmarshalProto(body))
		rms := payload.Metrics().ResourceMetrics()
		require.Equal(t, 1, rms.Len())
		attrs := rms.At(0).Resource().Attributes().AsRaw()
		require.Equal(t, "cp-test", attrs["nuon.control_plane.id"])
		require.Equal(t, "api-test", attrs["service.instance.id"])
		require.Equal(t, "ctl-api-test", attrs["service.name"])
		got := map[string]pmetric.Metric{}
		scopes := rms.At(0).ScopeMetrics()
		for i := 0; i < scopes.Len(); i++ {
			metrics := scopes.At(i).Metrics()
			for j := 0; j < metrics.Len(); j++ {
				metric := metrics.At(j)
				require.NotContains(t, got, metric.Name())
				got[metric.Name()] = metric
			}
		}
		require.Len(t, got, 20)
		for _, spec := range []struct {
			name, unit          string
			gauge, nonmonotonic bool
			values              []float64
		}{
			{"pgxpool.acquires", "", false, false, []float64{0, 0, 0}},
			{"pgxpool.acquire_duration", "ns", false, false, []float64{0, 0, 0}},
			{"pgxpool.acquired_connections", "", false, true, []float64{0, 0, 0}},
			{"pgxpool.canceled_acquires", "", false, false, []float64{float64(count), 11, 12}},
			{"pgxpool.constructing_connections", "", false, true, []float64{0, 0, 0}},
			{"pgxpool.empty_acquire", "", false, false, []float64{0, 0, 0}},
			{"pgxpool.idle_connections", "", false, true, []float64{0, 0, 0}},
			{"pgxpool.max_connections", "", true, false, []float64{17, 18, 19}},
			{"pgxpool.max_idle_destroys", "", false, false, []float64{0, 0, 0}},
			{"pgxpool.max_lifetime_destroys", "", false, false, []float64{0, 0, 0}},
			{"pgxpool.new_connections", "", false, false, []float64{0, 0, 0}},
			{"pgxpool.total_connections", "", false, true, []float64{0, 0, 0}},
			{"pgxpool.empty_acquire_wait_time", "ns", false, false, []float64{0, 0, 0}},
			{"db.sql.connection.max_open", "", true, false, []float64{3}},
			{"db.sql.connection.open", "", true, false, []float64{2, 1}},
			{"db.sql.connection.wait", "", false, false, []float64{1}},
			{"db.sql.connection.wait_duration", "ms", false, false, []float64{float64(waitDuration) / float64(time.Millisecond)}},
			{"db.sql.connection.closed_max_idle", "", false, false, []float64{0}},
			{"db.sql.connection.closed_max_idle_time", "", false, false, []float64{0}},
			{"db.sql.connection.closed_max_lifetime", "", false, false, []float64{0}},
		} {
			require.Contains(t, got, spec.name)
			metric := got[spec.name]
			require.Equal(t, spec.unit, metric.Unit(), spec.name)
			var points pmetric.NumberDataPointSlice
			if spec.gauge {
				require.Equal(t, pmetric.MetricTypeGauge, metric.Type(), spec.name)
				points = metric.Gauge().DataPoints()
			} else {
				require.Equal(t, pmetric.MetricTypeSum, metric.Type(), spec.name)
				require.Equal(t, !spec.nonmonotonic, metric.Sum().IsMonotonic(), spec.name)
				require.Equal(t, pmetric.AggregationTemporalityCumulative, metric.Sum().AggregationTemporality(), spec.name)
				points = metric.Sum().DataPoints()
			}
			require.Equal(t, len(spec.values), points.Len(), spec.name)
			seen := map[string]bool{}
			for i := 0; i < points.Len(); i++ {
				p := points.At(i)
				attributes := p.Attributes().AsRaw()
				role, ok := attributes["db.client.connection.pool.name"].(string)
				require.True(t, ok)
				wantAttrs := map[string]any{"db.client.connection.pool.name": role, "db.system.name": "postgresql"}
				index := slices.Index(roles, role)
				require.NotEqual(t, -1, index)
				key := role
				if strings.HasPrefix(spec.name, "db.sql.") {
					wantAttrs["db.system.name"] = "clickhouse"
					require.Equal(t, "primary", role)
					if spec.name == "db.sql.connection.open" {
						status, ok := attributes["status"].(string)
						require.True(t, ok)
						index = slices.Index([]string{"inuse", "idle"}, status)
						require.NotEqual(t, -1, index)
						wantAttrs["status"] = status
						key = status
					}
				}
				require.False(t, seen[key], spec.name)
				seen[key] = true
				require.Equal(t, wantAttrs, attributes, spec.name)
				value := p.DoubleValue()
				if p.ValueType() == pmetric.NumberDataPointValueTypeInt {
					value = float64(p.IntValue())
				}
				require.Equal(t, spec.values[index], value, spec.name)
			}
		}
	}
}

func TestPoolMetricsWithoutEndpoint(t *testing.T) {
	cfg, err := telemetry.NewConfig(&internal.Config{})
	require.NoError(t, err)
	lc := fxtest.NewLifecycle(t)
	provider, err := telemetry.NewMeterProvider(lc, cfg)
	require.NoError(t, err)
	require.IsType(t, noop.NewMeterProvider(), provider)
	m := New(provider)
	require.NoError(t, m.RegisterPostgres("primary", postgresPool(t, 17)))
	db := sql.OpenDB(poolConnector{})
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, m.RegisterClickHouse(lc, db))
	lc.RequireStart().RequireStop()
}

func TestClickHouseObserverLifecycle(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	lc := fxtest.NewLifecycle(t)
	db := sql.OpenDB(poolConnector{})
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, New(provider).RegisterClickHouse(lc, db))
	lc.RequireStart()
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	require.Len(t, data.ScopeMetrics, 1)
	require.Len(t, data.ScopeMetrics[0].Metrics, 7)
	for _, m := range data.ScopeMetrics[0].Metrics {
		if m.Name == "db.sql.connection.max_open" {
			points := m.Data.(metricdata.Gauge[int64]).DataPoints
			require.Len(t, points, 1)
			require.Zero(t, points[0].Value)
		}
	}
	lc.RequireStop()
	require.NoError(t, reader.Collect(context.Background(), &data))
	require.Empty(t, data.ScopeMetrics)
}
