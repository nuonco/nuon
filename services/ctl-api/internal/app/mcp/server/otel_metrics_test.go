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
	for _, path := range []string{"/", "/arbitrary-a?token=hidden", "/arbitrary-b"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("POST", path, nil))
		require.Equal(t, http.StatusUnauthorized, response.Code)
	}
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	found := false
	for _, scope := range data.ScopeMetrics {
		for _, metric := range scope.Metrics {
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
}
