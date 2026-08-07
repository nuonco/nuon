package tracer

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func setupTest(t *testing.T) (*gin.Engine, *tracetest.SpanRecorder) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware{}.Handler())

	return engine, recorder
}

func attrMap(span sdktrace.ReadOnlySpan) map[attribute.Key]attribute.Value {
	m := map[attribute.Key]attribute.Value{}
	for _, kv := range span.Attributes() {
		m[kv.Key] = kv.Value
	}
	return m
}

func TestRootSpanAttributes(t *testing.T) {
	engine, recorder := setupTest(t)

	engine.GET("/v1/apps/:app_id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/apps/app123", nil)
	engine.ServeHTTP(w, req)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	span := spans[0]

	require.Equal(t, "GET /v1/apps/:app_id", span.Name())
	require.Equal(t, oteltrace.SpanKindServer, span.SpanKind())

	attrs := attrMap(span)
	require.Equal(t, "GET", attrs["http.method"].AsString())
	require.Equal(t, "/v1/apps/:app_id", attrs["http.route"].AsString())
	require.Equal(t, int64(http.StatusOK), attrs["http.status_code"].AsInt64())
	require.NotEmpty(t, attrs["nuon.trace_id"].AsString())
	require.Equal(t, codes.Unset, span.Status().Code)
}

func TestDownstreamSpansParentToRootSpan(t *testing.T) {
	engine, recorder := setupTest(t)

	engine.GET("/v1/things", func(c *gin.Context) {
		_, child := otel.Tracer("test").Start(c.Request.Context(), "child.op")
		child.End()
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/things", nil)
	engine.ServeHTTP(w, req)

	spans := recorder.Ended()
	require.Len(t, spans, 2)

	var child, root sdktrace.ReadOnlySpan
	for _, s := range spans {
		switch s.Name() {
		case "child.op":
			child = s
		case "GET /v1/things":
			root = s
		}
	}
	require.NotNil(t, child)
	require.NotNil(t, root)
	require.Equal(t, root.SpanContext().TraceID(), child.SpanContext().TraceID())
	require.Equal(t, root.SpanContext().SpanID(), child.Parent().SpanID())
}

func TestServerErrorMarksSpanError(t *testing.T) {
	engine, recorder := setupTest(t)

	engine.GET("/v1/broken", func(c *gin.Context) {
		_ = c.Error(errors.New("boom"))
		c.Status(http.StatusInternalServerError)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/broken", nil)
	engine.ServeHTTP(w, req)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	span := spans[0]

	require.Equal(t, codes.Error, span.Status().Code)
	attrs := attrMap(span)
	require.Equal(t, int64(http.StatusInternalServerError), attrs["http.status_code"].AsInt64())

	var recorded bool
	for _, ev := range span.Events() {
		if ev.Name == "exception" {
			recorded = true
		}
	}
	require.True(t, recorded)
}

func TestClientErrorNotMarkedAsSpanError(t *testing.T) {
	engine, recorder := setupTest(t)

	engine.GET("/v1/things", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/things", nil)
	engine.ServeHTTP(w, req)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, codes.Unset, spans[0].Status().Code)
}

func TestMetricContextAttributes(t *testing.T) {
	engine, recorder := setupTest(t)

	engine.GET("/v1/things", func(c *gin.Context) {
		c.Set(keys.MetricsKey, &cctx.MetricContext{
			OrgID:     "org123",
			Namespace: "public",
			Context:   "api",
		})
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/things", nil)
	engine.ServeHTTP(w, req)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	attrs := attrMap(spans[0])
	require.Equal(t, "org123", attrs["nuon.org_id"].AsString())
	require.Equal(t, "public", attrs["nuon.namespace"].AsString())
	require.Equal(t, "api", attrs["nuon.context"].AsString())
}

func TestNuonTraceIDHeaderPreserved(t *testing.T) {
	engine, _ := setupTest(t)

	engine.GET("/v1/things", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("generated when missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/things", nil)
		engine.ServeHTTP(w, req)
		require.NotEmpty(t, w.Header().Get("X-Nuon-Trace-ID"))
	})

	t.Run("propagated when provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/things", nil)
		req.Header.Set("X-Nuon-Trace-ID", "incoming-trace-id")
		engine.ServeHTTP(w, req)
		require.Equal(t, "incoming-trace-id", w.Header().Get("X-Nuon-Trace-ID"))
	})
}

func TestUnmatchedRoute(t *testing.T) {
	engine, recorder := setupTest(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	engine.ServeHTTP(w, req)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "GET unmatched", spans[0].Name())
	attrs := attrMap(spans[0])
	require.Equal(t, "unmatched", attrs["http.route"].AsString())
}
