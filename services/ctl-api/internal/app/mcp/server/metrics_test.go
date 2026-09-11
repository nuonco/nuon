package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type capturedMetric struct {
	name  string
	tags  []string
	value float64
}

type capturingMetricsWriter struct {
	mu      sync.Mutex
	gauges  []capturedMetric
	timings []capturedMetric
}

func (*capturingMetricsWriter) Incr(string, []string)                  {}
func (*capturingMetricsWriter) Decr(string, []string)                  {}
func (*capturingMetricsWriter) Count(string, int64, []string)          {}
func (*capturingMetricsWriter) Distribution(string, float64, []string) {}
func (*capturingMetricsWriter) Event(*statsd.Event)                    {}
func (*capturingMetricsWriter) Flush()                                 {}

func (w *capturingMetricsWriter) Gauge(name string, value float64, tags []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.gauges = append(w.gauges, capturedMetric{name: name, value: value, tags: append([]string(nil), tags...)})
}

func (w *capturingMetricsWriter) Timing(name string, value time.Duration, tags []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.timings = append(w.timings, capturedMetric{name: name, value: float64(value), tags: append([]string(nil), tags...)})
}

func (w *capturingMetricsWriter) timingForEndpoint(endpoint string) (capturedMetric, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, metric := range w.timings {
		if metric.name == "api.request.latency" && hasTag(metric.tags, "endpoint:"+endpoint) {
			return metric, true
		}
	}
	return capturedMetric{}, false
}

func hasTag(tags []string, want string) bool {
	for _, tag := range tags {
		if tag == want {
			return true
		}
	}
	return false
}

type failingMCPService struct {
	stubMCPService
}

func (failingMCPService) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, api.MCPReadTool(
		"fail",
		"Fail",
		"return a tool error",
	), func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
		return nil, nil, errors.New("tool failed")
	})
}

type metricContextMCPService struct {
	stubMCPService
}

func (metricContextMCPService) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, api.MCPReadTool(
		"metric_context",
		"Metric context",
		"return the request metric context",
	), func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		metricCtx, err := cctx.MetricsContextFromGinContext(ctx)
		if err != nil {
			return nil, nil, err
		}
		return api.MCPJSONResult(map[string]string{
			"endpoint": metricCtx.Endpoint,
			"org_id":   metricCtx.OrgID,
		})
	})
}

func newMetricsTestServer(t *testing.T, service api.MCPService, writer *capturingMetricsWriter) *httptest.Server {
	t.Helper()

	s := &Server{
		mw:                 writer,
		mcpServices:        []api.MCPService{service},
		schemaCache:        mcp.NewSchemaCache(),
		implementationName: "nuon-ctl",
		orgSelections:      make(map[string]*orgSelection),
	}
	mcpHandler := s.newMCPHandler()
	handler := s.metricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := withMCPAuth(r.Context(), "org_a", "acc_test")
		mcpHandler.ServeHTTP(w, r.WithContext(ctx))
	}))
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	return ts
}

func TestMCPMetricsTagsToolCalls(t *testing.T) {
	writer := &capturingMetricsWriter{}
	ts := newMetricsTestServer(t, stubMCPService{}, writer)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	require.NoError(t, err)
	defer session.Close()

	_, err = session.CallTool(context.Background(), &mcp.CallToolParams{Name: "echo_org"})
	require.NoError(t, err)

	metric, ok := writer.timingForEndpoint("tools/call/echo_org")
	require.True(t, ok)
	assert.True(t, hasTag(metric.tags, "context:mcp_api"))
	assert.True(t, hasTag(metric.tags, "method:POST"))
	assert.True(t, hasTag(metric.tags, "status:ok"))
	assert.True(t, hasTag(metric.tags, "status_code_class:2xx"))
	assert.True(t, hasTag(metric.tags, "is_panic:false"))
	assert.True(t, hasTag(metric.tags, "is_timeout:false"))
	assert.True(t, hasTag(metric.tags, "is_deprecated:false"))
}

func TestMCPMetricsMarksToolErrors(t *testing.T) {
	writer := &capturingMetricsWriter{}
	ts := newMetricsTestServer(t, failingMCPService{}, writer)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	require.NoError(t, err)
	defer session.Close()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "fail"})
	require.NoError(t, err)
	require.True(t, result.IsError)

	metric, ok := writer.timingForEndpoint("tools/call/fail")
	require.True(t, ok)
	assert.True(t, hasTag(metric.tags, "status:err"))
	assert.True(t, hasTag(metric.tags, "status_code_class:2xx"))
}

func TestMCPMetricsPopulatesDatabaseMetricContext(t *testing.T) {
	writer := &capturingMetricsWriter{}
	ts := newMetricsTestServer(t, metricContextMCPService{}, writer)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	require.NoError(t, err)
	defer session.Close()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "metric_context"})
	require.NoError(t, err)
	require.Len(t, result.Content, 1)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.JSONEq(t, `{"endpoint":"tools/call/metric_context","org_id":"org_a"}`, text.Text)
}

func TestMCPMetricsSkipsHealthChecks(t *testing.T) {
	writer := &capturingMetricsWriter{}
	handler := (&Server{mw: writer}).metricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/livez", nil))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/readyz", nil))

	assert.Empty(t, writer.timings)
	assert.Empty(t, writer.gauges)
}

func TestMCPMetricsRecordsPanics(t *testing.T) {
	writer := &capturingMetricsWriter{}
	handler := (&Server{mw: writer}).metricsMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	}))

	assert.Panics(t, func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
	})

	metric, ok := writer.timingForEndpoint("unauthenticated")
	require.True(t, ok)
	assert.True(t, hasTag(metric.tags, "status:err"))
	assert.True(t, hasTag(metric.tags, "status_code_class:5xx"))
	assert.True(t, hasTag(metric.tags, "is_panic:true"))
}
