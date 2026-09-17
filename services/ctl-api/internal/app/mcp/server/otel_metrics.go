package server

import (
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
)

func (s *Server) otelMetricsMiddleware(next http.Handler) http.Handler {
	if s.httpMetrics == nil {
		return next
	}
	return s.httpMetrics.Handler("mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := "/"
		switch r.URL.Path {
		case "/livez", "/readyz", "/.well-known/oauth-protected-resource":
			route = r.URL.Path
		}
		metrics.SetHTTPRoute(r.Context(), route)
		next.ServeHTTP(w, r)
	}))
}
