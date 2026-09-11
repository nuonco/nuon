package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
)

func TestOTELMetricsIncludesAuthFailuresWithoutRawPaths(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	m, err := metrics.NewHTTPMetrics(provider)
	require.NoError(t, err)
	s := &Server{httpMetrics: m, l: zap.NewNop()}
	handler := s.otelMetricsMiddleware(s.authContextMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("unauthenticated request reached handler")
	})))
	for i, path := range []string{"/", "/arbitrary-a?token=hidden", "/arbitrary-b"} {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("POST", path, nil)
		req.ContentLength = int64(i+1) * 17
		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusUnauthorized, response.Code)
	}
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	found := false
	foundSize := false
	for _, scope := range data.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name == "nuon.http.server.request.declared_body.size" {
				foundSize = true
				points := metric.Data.(metricdata.Histogram[int64]).DataPoints
				require.Len(t, points, 1)
				require.EqualValues(t, 3, points[0].Count)
				require.EqualValues(t, 102, points[0].Sum)
				status, _ := points[0].Attributes.Value("http.response.status_code")
				require.EqualValues(t, 401, status.AsInt64())
			}
			if metric.Name == "http.server.request.duration" {
				found = true
				points := metric.Data.(metricdata.Histogram[float64]).DataPoints
				require.Len(t, points, 1)
				require.EqualValues(t, 3, points[0].Count)
				route, _ := points[0].Attributes.Value("http.route")
				require.Equal(t, "/", route.AsString())
				api, _ := points[0].Attributes.Value("nuon.api")
				require.Equal(t, "mcp", api.AsString())
			}
		}
	}
	require.True(t, found)
	require.True(t, foundSize)
}

func TestOTELBodySizeOnPanicAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		path   string
		panic  bool
		cancel bool
	}{
		{"panic", "/", true, false},
		{"canceled", "/", false, true},
		{"health", "/readyz", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := sdkmetric.NewManualReader()
			provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
			m, err := metrics.NewHTTPMetrics(provider)
			require.NoError(t, err)
			s := &Server{httpMetrics: m}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			handler := s.otelMetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.panic {
					panic("test")
				}
				if tc.cancel {
					cancel()
				}
				w.WriteHeader(http.StatusAccepted)
			}))
			req := httptest.NewRequest("POST", tc.path, nil).WithContext(ctx)
			req.ContentLength = 73
			run := func() { handler.ServeHTTP(httptest.NewRecorder(), req) }
			if tc.panic {
				require.Panics(t, run)
			} else {
				run()
			}
			var data metricdata.ResourceMetrics
			require.NoError(t, reader.Collect(context.Background(), &data))
			require.Len(t, data.ScopeMetrics, 1)
			require.Len(t, data.ScopeMetrics[0].Metrics, 3)
			found := map[string]int64{}
			for _, metric := range data.ScopeMetrics[0].Metrics {
				switch metric.Name {
				case "nuon.http.server.request.declared_body.size":
					points := metric.Data.(metricdata.Histogram[int64]).DataPoints
					require.Len(t, points, 1)
					require.EqualValues(t, 1, points[0].Count)
					status, _ := points[0].Attributes.Value("http.response.status_code")
					if tc.panic {
						require.EqualValues(t, 500, status.AsInt64())
					} else {
						require.EqualValues(t, 202, status.AsInt64())
					}
					found[metric.Name] = points[0].Sum
				case "http.server.active_requests":
					require.Zero(t, metric.Data.(metricdata.Sum[int64]).DataPoints[0].Value)
				}
			}
			want := map[string]int64{"nuon.http.server.request.declared_body.size": 73}
			require.Equal(t, want, found)
		})
	}
}
