package metrics

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	require.Len(t, data.ScopeMetrics[0].Metrics, 3)
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

func TestHTTPDeclaredBodySize(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	m, err := NewHTTPMetrics(provider)
	require.NoError(t, err)
	for i, declared := range []int64{257, 0, 4097, -1} {
		body := strings.NewReader("payload")
		req := httptest.NewRequest("POST", fmt.Sprintf("/items/%d?token=hidden", i), nil)
		req.Body = io.NopCloser(body)
		req.ContentLength = declared
		handler := m.Handler("public", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, req.Body, r.Body, "instrumentation must not wrap the body")
			require.Equal(t, 7, body.Len(), "instrumentation must not read the body")
			SetHTTPRoute(r.Context(), "/items/:id")
			if declared != 0 {
				_, err := io.ReadFull(r.Body, make([]byte, 2))
				require.NoError(t, err)
			}
			r.ContentLength = 999
			w.WriteHeader(http.StatusCreated)
		}))
		handler.ServeHTTP(httptest.NewRecorder(), req)
		if declared == 0 {
			require.Equal(t, 7, body.Len())
		} else {
			require.Equal(t, 5, body.Len())
		}
	}
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	require.Len(t, data.ScopeMetrics, 1)
	require.Len(t, data.ScopeMetrics[0].Metrics, 3)
	for _, metric := range data.ScopeMetrics[0].Metrics {
		if metric.Name == "http.server.request.duration" {
			require.EqualValues(t, 4, metric.Data.(metricdata.Histogram[float64]).DataPoints[0].Count)
			continue
		}
		if metric.Name == "http.server.active_requests" {
			require.Zero(t, metric.Data.(metricdata.Sum[int64]).DataPoints[0].Value)
			continue
		}
		histogram := metric.Data.(metricdata.Histogram[int64])
		require.Equal(t, metricdata.CumulativeTemporality, histogram.Temporality)
		require.Len(t, histogram.DataPoints, 1)
		point := histogram.DataPoints[0]
		require.EqualValues(t, 3, point.Count)
		require.Equal(t, attribute.NewSet(
			attribute.String("nuon.api", "public"),
			attribute.String("http.request.method", "POST"),
			attribute.String("url.scheme", "http"),
			attribute.String("http.route", "/items/:id"),
			attribute.Int("http.response.status_code", 201),
		), point.Attributes)
		require.Equal(t, "nuon.http.server.request.declared_body.size", metric.Name)
		require.Equal(t, "By", metric.Unit)
		require.EqualValues(t, 4354, point.Sum)
		require.Equal(t, []float64{0, 128, 512, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216}, point.Bounds)
		require.Equal(t, []uint64{1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0}, point.BucketCounts)
	}
}
