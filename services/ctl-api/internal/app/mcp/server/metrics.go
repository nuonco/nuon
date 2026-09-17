package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

const (
	mcpMetricsContext = "mcp_api"
	mcpTargetLatency  = 50 * time.Millisecond
)

type mcpCallStatus struct {
	isError bool
}

func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.mw == nil || r.URL.Path == "/livez" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}

		startTS := time.Now()
		metricCtx := &cctx.MetricContext{
			RequestURI: r.URL.RequestURI(),
			Endpoint:   mcpEndpointFromRequest(r),
			Method:     r.Method,
			Context:    mcpMetricsContext,
		}
		callStatus := &mcpCallStatus{}
		ctx := context.WithValue(r.Context(), keys.MetricsKey, metricCtx)
		ctx = context.WithValue(ctx, mcpCallStatusKey{}, callStatus)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			recovered := recover()
			if recovered != nil {
				metricCtx.IsPanic = true
				rec.status = http.StatusInternalServerError
			}
			if r.Context().Err() == context.DeadlineExceeded {
				metricCtx.IsTimeout = true
			}

			status := "ok"
			if rec.status >= 400 || callStatus.isError {
				status = "err"
			}

			duration := time.Since(startTS)
			tags := []string{
				"status:" + status,
				"status_code_class:" + fmt.Sprintf("%dxx", rec.status/100),
				"endpoint:" + metricCtx.Endpoint,
				"method:" + metricCtx.Method,
				"context:" + mcpMetricsContext,
				"within_target_latency:" + strconv.FormatBool(duration < mcpTargetLatency),
				"is_panic:" + strconv.FormatBool(metricCtx.IsPanic),
				"is_timeout:" + strconv.FormatBool(metricCtx.IsTimeout),
				"is_deprecated:" + strconv.FormatBool(metricCtx.IsDeprecated),
			}

			s.mw.Gauge("api.request.size", float64(r.ContentLength), tags)
			s.mw.Timing("api.request.latency", duration, tags)
			s.mw.Gauge("gorm_operation.endpoint_count", float64(metricCtx.DBQueryCount), tags)

			if recovered != nil {
				panic(recovered)
			}
		}()

		next.ServeHTTP(rec, r.WithContext(ctx))
	})
}

func (s *Server) receivingMetricsMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		if metricCtx, err := cctx.MetricsContextFromGinContext(ctx); err == nil {
			toolName := mcpToolName(req)
			metricCtx.Endpoint = mcpRPCEndpoint(method, toolName)
			metricCtx.OrgID = keys.OrgIDFromContext(ctx)
		}

		res, err := next(ctx, method, req)
		if callStatus, ok := ctx.Value(mcpCallStatusKey{}).(*mcpCallStatus); ok {
			if err != nil {
				callStatus.isError = true
			}
			if cr, ok := res.(*mcp.CallToolResult); ok && cr != nil && cr.IsError {
				callStatus.isError = true
			}
		}
		return res, err
	}
}

type mcpCallStatusKey struct{}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(p)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func mcpEndpointFromRequest(r *http.Request) string {
	if r.URL.Path == "/" || r.URL.Path == "" {
		return "unauthenticated"
	}
	return strings.ReplaceAll(r.URL.Path, "-", "_")
}

func mcpRPCEndpoint(method, toolName string) string {
	endpoint := strings.ReplaceAll(method, "-", "_")
	if method == "tools/call" && toolName != "" {
		return endpoint + "/" + strings.ReplaceAll(toolName, "-", "_")
	}
	return endpoint
}

func mcpToolName(req mcp.Request) string {
	switch p := req.GetParams().(type) {
	case *mcp.CallToolParamsRaw:
		return p.Name
	case *mcp.CallToolParams:
		return p.Name
	default:
		return ""
	}
}
