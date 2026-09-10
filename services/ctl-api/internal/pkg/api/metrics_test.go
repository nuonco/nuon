package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

func TestAPIMetricsAllSurfaces(t *testing.T) {
	constructors := map[string]func(Params) (*API, error){
		"public": NewPublicAPI, "internal": NewInternalAPI, "runner": NewRunnerAPI,
		"auth": NewAuthAPI, "admin-dashboard": NewAdminDashboardAPI, "slack": NewSlackAPI,
	}
	for surface, constructor := range constructors {
		t.Run(surface, func(t *testing.T) {
			reader := sdkmetric.NewManualReader()
			provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
			m, err := metrics.NewHTTPMetrics(provider)
			require.NoError(t, err)
			a, err := constructor(Params{
				Cfg: &internal.Config{}, L: zap.NewNop(), LC: fxtest.NewLifecycle(t), HTTPMetrics: m,
			})
			require.NoError(t, err)
			a.handler.Use(gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) {
				c.AbortWithStatus(http.StatusInternalServerError)
			}))
			a.handler.Use(func(c *gin.Context) {
				if c.Request.Method == "POST" {
					c.AbortWithStatus(http.StatusUnauthorized)
				}
			})
			a.handler.GET("/v1/items/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
			a.handler.POST("/v1/items/:id", func(c *gin.Context) { t.Fatal("auth failure reached handler") })
			a.handler.GET("/panic", func(c *gin.Context) { panic("test") })
			a.handler.GET("/readyz", func(c *gin.Context) { c.Status(http.StatusOK) })
			requests := []struct {
				method, path string
				status       int
			}{
				{"GET", "/v1/items/a?token=hidden", 200},
				{"GET", "/v1/items/b", 200},
				{"POST", "/v1/items/a", 401},
				{"GET", "/panic", 500},
				{"GET", "/missing-a", 404},
				{"GET", "/missing-b", 404},
				{"GET", "/v1/items/a/", 301},
				{"CUSTOM_A", "/missing-a", 404},
				{"CUSTOM_B", "/missing-b", 404},
				{"GET", "/readyz", 200},
			}
			for _, req := range requests {
				response := httptest.NewRecorder()
				a.srv.Handler.ServeHTTP(response, httptest.NewRequest(req.method, req.path, nil))
				require.Equal(t, req.status, response.Code)
			}
			var data metricdata.ResourceMetrics
			require.NoError(t, reader.Collect(context.Background(), &data))
			observed := map[string]uint64{}
			for _, scope := range data.ScopeMetrics {
				for _, metric := range scope.Metrics {
					if metric.Name != "http.server.request.duration" {
						continue
					}
					for _, point := range metric.Data.(metricdata.Histogram[float64]).DataPoints {
						api, _ := point.Attributes.Value("nuon.api")
						require.Equal(t, surface, api.AsString())
						method, _ := point.Attributes.Value("http.request.method")
						route, _ := point.Attributes.Value("http.route")
						status, _ := point.Attributes.Value("http.response.status_code")
						errorType, hasError := point.Attributes.Value("error.type")
						require.Equal(t, status.AsInt64() >= 500, hasError)
						if hasError {
							require.Equal(t, "500", errorType.AsString())
						}
						observed[fmt.Sprintf("%s|%s|%d", method.AsString(), route.AsString(), status.AsInt64())] += point.Count
					}
				}
			}
			require.Equal(t, map[string]uint64{
				"GET|/v1/items/:id|200":  2,
				"POST|/v1/items/:id|401": 1,
				"GET|/panic|500":         1,
				"GET||404":               2,
				"GET||301":               1,
				"_OTHER||404":            2,
				"GET|/readyz|200":        1,
			}, observed)
		})
	}
}

func TestFXShutdownDrainsAPIBeforeMetricExport(t *testing.T) {
	t.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "3600000")
	requests := make(chan []byte, 4)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- body
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	defer receiver.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	defer unblock()
	var a *API
	var testServer *httptest.Server
	app := fx.New(fx.NopLogger,
		fx.Supply(&internal.Config{OTELExporterOTLPEndpoint: receiver.URL, GracefulShutdownTimeout: time.Second}),
		fx.Provide(telemetry.NewConfig, telemetry.NewMeterProvider, metrics.NewHTTPMetrics),
		fx.Invoke(func(lc fx.Lifecycle, cfg *internal.Config, m *metrics.HTTPMetrics) {
			a = &API{cfg: cfg, l: zap.NewNop(), name: "public"}
			require.NoError(t, a.init(m))
			a.handler.GET("/drain", func(c *gin.Context) {
				close(entered)
				<-release
				c.Status(http.StatusAccepted)
			})
			testServer = httptest.NewUnstartedServer(a.srv.Handler)
			a.srv = testServer.Config
			hooks := a.lifecycleHooks(nil)
			// httptest supplies a reserved ephemeral listener; use the real API stop hook.
			hooks.OnStart = func(context.Context) error { testServer.Start(); return nil }
			lc.Append(hooks)
		}),
	)
	require.NoError(t, app.Start(context.Background()))
	defer func() { unblock(); testServer.Close() }()
	draining := make(chan struct{})
	a.srv.RegisterOnShutdown(func() { close(draining) })
	response := make(chan int, 1)
	go func() {
		r, err := testServer.Client().Get(testServer.URL + "/drain")
		if err != nil {
			response <- 0
			return
		}
		defer r.Body.Close()
		_, _ = io.Copy(io.Discard, r.Body)
		response <- r.StatusCode
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	stopped := make(chan error, 1)
	go func() { stopped <- app.Stop(context.Background()) }()
	select {
	case <-draining:
	case <-time.After(time.Second):
		t.Fatal("API shutdown did not start")
	}
	unblock()
	require.NoError(t, <-stopped)
	require.Equal(t, http.StatusAccepted, <-response)
	payload := pmetricotlp.NewExportRequest()
	select {
	case body := <-requests:
		require.NoError(t, payload.UnmarshalProto(body))
	case <-time.After(time.Second):
		t.Fatal("no shutdown export received")
	}
	observed := payload.Metrics().ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	found := false
	for i := 0; i < observed.Len(); i++ {
		m := observed.At(i)
		if m.Name() == "http.server.request.duration" {
			found = true
			require.Equal(t, 1, m.Histogram().DataPoints().Len())
			require.EqualValues(t, 1, m.Histogram().DataPoints().At(0).Count())
			status, _ := m.Histogram().DataPoints().At(0).Attributes().Get("http.response.status_code")
			require.EqualValues(t, http.StatusAccepted, status.Int())
		}
		if m.Name() == "http.server.active_requests" {
			require.Zero(t, m.Sum().DataPoints().At(0).IntValue())
		}
	}
	require.True(t, found, "the final export must include the drained request")
}
