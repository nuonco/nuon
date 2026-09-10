package health

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type healthDB struct {
	t                          *testing.T
	pingErr, queryErr, rowsErr error
	values                     []driver.Value
	pings, queries             int
}

type healthDriver struct{}
type healthConn struct{ db *healthDB }
type healthRows struct {
	values []driver.Value
	err    error
}

func (d *healthDB) Connect(context.Context) (driver.Conn, error) { return healthConn{d}, nil }
func (d *healthDB) Driver() driver.Driver                        { return healthDriver{} }
func (healthDriver) Open(string) (driver.Conn, error)            { return nil, errors.New("unexpected open") }
func (healthConn) Close() error                                  { return nil }
func (healthConn) Begin() (driver.Tx, error)                     { return nil, errors.New("unexpected transaction") }
func (healthConn) Prepare(string) (driver.Stmt, error)           { return nil, errors.New("unexpected prepare") }
func (c healthConn) Ping(context.Context) error {
	c.db.pings++
	return c.db.pingErr
}
func (c healthConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.db.queries++
	require.Equal(c.db.t, "SELECT table FROM system.replicas WHERE database = 'ctl_api' AND is_readonly = 1", query)
	require.Empty(c.db.t, args)
	return &healthRows{values: c.db.values, err: c.db.rowsErr}, c.db.queryErr
}
func (healthRows) Columns() []string { return []string{"table"} }
func (healthRows) Close() error      { return nil }
func (r *healthRows) Next(dest []driver.Value) error {
	if len(r.values) == 0 {
		if r.err != nil {
			return r.err
		}
		return io.EOF
	}
	dest[0], r.values = r.values[0], r.values[1:]
	return nil
}

func TestReadyzDependencyMetrics(t *testing.T) {
	failure := errors.New("private database failure detail")
	for _, tt := range []struct {
		name                        string
		pgConnection, chConnection  bool
		pgPing, chPing, query, rows error
		values                      []driver.Value
		temporal                    error
		cancel, panic               bool
		status                      int
		degraded                    []string
		reasons                     [dependencyCount]string
	}{
		{name: "healthy", status: 200, degraded: []string{}},
		{name: "postgres_connection", pgConnection: true, status: 500, reasons: [3]string{"connection", "skipped", "skipped"}},
		{name: "postgres_ping", pgPing: failure, status: 500, reasons: [3]string{"ping", "skipped", "skipped"}},
		{name: "canceled_request", cancel: true, pgPing: context.Canceled, status: 500, reasons: [3]string{"ping", "skipped", "skipped"}},
		{name: "clickhouse_connection", chConnection: true, status: 207, degraded: []string{"ch"}, reasons: [3]string{"", "connection", ""}},
		{name: "clickhouse_ping", chPing: failure, status: 207, degraded: []string{"ch"}, reasons: [3]string{"", "ping", ""}},
		{name: "clickhouse_query", query: failure, status: 207, degraded: []string{"ch"}, reasons: [3]string{"", "query", ""}},
		{name: "clickhouse_scan", values: []driver.Value{nil}, status: 207, degraded: []string{"ch"}, reasons: [3]string{"", "scan", ""}},
		{name: "clickhouse_iteration", rows: failure, status: 207, degraded: []string{"ch"}, reasons: [3]string{"", "iteration", ""}},
		{name: "readonly_replicas", values: []driver.Value{"events", "jobs"}, status: 207, degraded: []string{"ch"}, reasons: [3]string{"", "readonly_replicas", ""}},
		{name: "multiple_clickhouse_failures", chPing: failure, query: failure, status: 207, degraded: []string{"ch", "ch"}, reasons: [3]string{"", "query", ""}},
		{name: "temporal", temporal: failure, status: 207, degraded: []string{"temporal"}, reasons: [3]string{"", "", "ping"}},
		{name: "temporal_panic", panic: true, status: 500, reasons: [3]string{"", "", "incomplete"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reader := sdkmetric.NewManualReader()
			provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
			pg := &healthDB{t: t, pingErr: tt.pgPing}
			ch := &healthDB{t: t, pingErr: tt.chPing, queryErr: tt.query, rowsErr: tt.rows, values: tt.values}
			database := func(d *healthDB, connectionError bool) *gorm.DB {
				if connectionError {
					return &gorm.DB{Config: &gorm.Config{}}
				}
				db := sql.OpenDB(d)
				t.Cleanup(func() { require.NoError(t, db.Close()) })
				return &gorm.DB{Config: &gorm.Config{ConnPool: db}}
			}
			ctrl := gomock.NewController(t)
			tc := temporalclient.NewMockClient(ctrl)
			if tt.reasons[temporalDependency] != "skipped" {
				tc.EXPECT().CheckHealth(gomock.Any(), gomock.Any()).DoAndReturn(func(context.Context, *client.CheckHealthRequest) (*client.CheckHealthResponse, error) {
					if tt.panic {
						panic("test health panic")
					}
					return &client.CheckHealthResponse{}, tt.temporal
				})
			}
			mw := metrics.NewMockWriter(ctrl)
			mw.EXPECT().Incr("healthcheck.check", gomock.Any()).AnyTimes()
			s, err := New(Params{DB: database(pg, tt.pgConnection), CHDB: database(ch, tt.chConnection), TClient: tc, MW: mw, MeterProvider: provider})
			require.NoError(t, err)
			router := tests.NewTestRouter(tests.RouterOptions{L: zap.NewNop(), DB: s.db})
			require.NoError(t, s.RegisterPublicRoutes(router))
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			if tt.cancel {
				ctx, cancel := context.WithCancel(request.Context())
				cancel()
				request = request.WithContext(ctx)
			}
			started := time.Now()
			if tt.panic {
				require.Panics(t, func() { router.ServeHTTP(response, request) })
			} else {
				router.ServeHTTP(response, request)
				require.Equal(t, tt.status, response.Code)
			}
			finished := time.Now()
			if tt.status != 500 {
				var body struct {
					Status   string   `json:"status"`
					Degraded []string `json:"degraded"`
				}
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
				require.Equal(t, tt.degraded, body.Degraded)
				if tt.status == 207 {
					require.Equal(t, "degraded", body.Status)
				} else {
					require.Equal(t, "ok", body.Status)
				}
			}
			if tt.name == "readonly_replicas" {
				require.Equal(t, "events", response.Header().Get("x-ch-table-in-read-only-0"))
				require.Equal(t, "jobs", response.Header().Get("x-ch-table-in-read-only-1"))
			}
			data := collectHealthMetrics(t, reader)
			counts := data["nuon.dependency.checks"].Data.(metricdata.Sum[int64]).DataPoints
			require.Len(t, counts, 3)
			for _, p := range counts {
				dep, _ := p.Attributes.Value("dependency.name")
				i := slices.Index([]string{"postgresql", "clickhouse", "temporal"}, dep.AsString())
				require.NotEqual(t, -1, i)
				reason := tt.reasons[i]
				attrs := []attribute.KeyValue{attribute.String("dependency.name", dep.AsString())}
				outcome := "success"
				if reason == "skipped" {
					outcome, reason = "skipped", "previous_dependency_failed"
				} else if reason != "" {
					outcome = "failure"
				}
				attrs = append(attrs, attribute.String("outcome", outcome))
				if reason != "" {
					attrs = append(attrs, attribute.String("error.type", reason))
				}
				require.Equal(t, attribute.NewSet(attrs...), p.Attributes)
				require.EqualValues(t, 1, p.Value, "one outcome per dependency, not per subcheck")
			}
			attempted := 0
			for _, reason := range tt.reasons {
				if reason != "skipped" {
					attempted++
				}
			}
			states := data["nuon.dependency.check.status"].Data.(metricdata.Gauge[int64]).DataPoints
			require.Len(t, states, attempted)
			for _, p := range states {
				dep, _ := p.Attributes.Value("dependency.name")
				i := slices.Index([]string{"postgresql", "clickhouse", "temporal"}, dep.AsString())
				want := int64(0)
				if tt.reasons[i] == "" {
					want = 1
				}
				require.Equal(t, want, p.Value)
			}
			durations := data["nuon.dependency.check.duration"].Data.(metricdata.Histogram[float64]).DataPoints
			require.Len(t, durations, attempted)
			for _, p := range durations {
				require.EqualValues(t, 1, p.Count)
				require.GreaterOrEqual(t, p.Sum, float64(0))
				require.LessOrEqual(t, p.Sum, finished.Sub(started).Seconds())
			}
			for _, p := range data["nuon.dependency.check.last_completed"].Data.(metricdata.Gauge[float64]).DataPoints {
				require.GreaterOrEqual(t, p.Value, float64(started.UnixNano())/1e9)
				require.LessOrEqual(t, p.Value, float64(finished.UnixNano())/1e9)
			}
			if tt.reasons[clickhouseDependency] == "skipped" || tt.chConnection {
				require.Zero(t, ch.pings)
				require.Zero(t, ch.queries)
			} else {
				require.Equal(t, 1, ch.pings)
				require.Equal(t, 1, ch.queries)
			}
			beforePings, beforeQueries := ch.pings, ch.queries
			collectHealthMetrics(t, reader)
			require.Equal(t, beforePings, ch.pings)
			require.Equal(t, beforeQueries, ch.queries)
		})
	}
}
