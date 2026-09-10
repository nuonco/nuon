package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestHTTPMetricsActiveAndStreaming(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	m, err := NewHTTPMetrics(provider)
	require.NoError(t, err)
	collect := func() metricdata.ResourceMetrics {
		var data metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(context.Background(), &data))
		return data
	}
	handler := m.Handler("auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetHTTPRoute(r.Context(), "/oauth/token")
		data := collect()
		require.Len(t, data.ScopeMetrics, 1)
		require.Len(t, data.ScopeMetrics[0].Metrics, 1)
		require.Equal(t, "http.server.active_requests", data.ScopeMetrics[0].Metrics[0].Name)
		for _, metric := range data.ScopeMetrics[0].Metrics {
			if metric.Name == "http.server.active_requests" {
				points := metric.Data.(metricdata.Sum[int64]).DataPoints
				require.Len(t, points, 1)
				require.EqualValues(t, 1, points[0].Value)
				require.Equal(t, attribute.NewSet(
					attribute.String("nuon.api", "auth"),
					attribute.String("http.request.method", "POST"),
					attribute.String("url.scheme", "https"),
				), points[0].Attributes)
			}
		}
		time.Sleep(20 * time.Millisecond)
		w.(http.Flusher).Flush()
		_, err := w.Write([]byte("streamed"))
		require.NoError(t, err)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("POST", "https://example.com/oauth/token?secret=hidden", nil))
	require.Equal(t, "streamed", response.Body.String())
	require.True(t, response.Flushed)
	data := collect()
	require.Len(t, data.ScopeMetrics, 1)
	require.Len(t, data.ScopeMetrics[0].Metrics, 2)
	for _, metric := range data.ScopeMetrics[0].Metrics {
		switch metric.Name {
		case "http.server.active_requests":
			require.Zero(t, metric.Data.(metricdata.Sum[int64]).DataPoints[0].Value)
		case "http.server.request.duration":
			point := metric.Data.(metricdata.Histogram[float64]).DataPoints[0]
			require.EqualValues(t, 1, point.Count)
			require.GreaterOrEqual(t, point.Sum, .02)
			require.Less(t, point.Sum, 10.0)
			require.Equal(t, []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10}, point.Bounds)
			require.Equal(t, attribute.NewSet(
				attribute.String("nuon.api", "auth"),
				attribute.String("http.request.method", "POST"),
				attribute.String("url.scheme", "https"),
				attribute.String("http.route", "/oauth/token"),
				attribute.Int("http.response.status_code", 200),
			), point.Attributes)
		}
	}
}
