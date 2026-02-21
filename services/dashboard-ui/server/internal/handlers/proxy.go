package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type ProxyHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewProxyHandler(cfg *internal.Config, l *zap.Logger) *ProxyHandler {
	return &ProxyHandler{cfg: cfg, l: l}
}

func (h *ProxyHandler) RegisterRoutes(e *gin.Engine) error {
	// Temporal UI proxy
	e.Any("/admin/temporal/*path", h.TemporalUIProxy)
	e.Any("/_app/*path", h.TemporalUIProxy)

	// ctl-api proxy — strips /api/ctl-api prefix and adds auth header
	e.Any("/api/ctl-api/*path", h.CtlAPIProxy)
	e.Any("/public/*path", h.CtlAPIDocsProxy)

	// Admin ctl-api proxy
	e.Any("/api/admin/ctl-api/*path", h.AdminCtlAPIProxy)
	e.Any("/admin/swagger/*path", h.AdminCtlAPIProxy)

	// API health check — proxied to ctl-api /v1/livez and wrapped in TAPIResponse
	e.GET("/api/livez", h.APILivez)

	return nil
}

func (h *ProxyHandler) reverseProxy(target string) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		h.l.Error("failed to parse proxy target", zap.String("target", target), zap.Error(err))
		return nil
	}
	return httputil.NewSingleHostReverseProxy(targetURL)
}

func (h *ProxyHandler) TemporalUIProxy(c *gin.Context) {
	proxy := h.reverseProxy(h.cfg.TemporalUIURL)
	if proxy == nil {
		c.Status(http.StatusBadGateway)
		return
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *ProxyHandler) CtlAPIProxy(c *gin.Context) {
	proxy := h.reverseProxy(h.cfg.NuonAPIURL)
	if proxy == nil {
		c.Status(http.StatusBadGateway)
		return
	}

	// Strip /api/ctl-api prefix so ctl-api sees /v1/...
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = strings.TrimPrefix(originalPath, "/api/ctl-api")
	if c.Request.URL.RawPath != "" {
		c.Request.URL.RawPath = strings.TrimPrefix(c.Request.URL.RawPath, "/api/ctl-api")
	}

	// Add Authorization header from the validated token stored by auth middleware
	if token, _ := cctx.TokenFromGinContext(c); token != "" {
		c.Request.Header.Set("Authorization", "Bearer "+token)
	}

	// Extract org ID from the path (e.g. /v1/orgs/<orgId>/...) and set as header.
	// The ctl-api requires X-Nuon-Org-ID for org-scoped endpoints.
	if orgID := extractOrgIDFromPath(c.Request.URL.Path); orgID != "" {
		c.Request.Header.Set("X-Nuon-Org-ID", orgID)
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// extractOrgIDFromPath extracts the org ID from paths like /v1/orgs/<orgId> or /v1/orgs/<orgId>/...
func extractOrgIDFromPath(path string) string {
	// Look for /v1/orgs/<orgId> pattern
	const prefix = "/v1/orgs/"
	idx := strings.Index(path, prefix)
	if idx < 0 {
		return ""
	}
	rest := path[idx+len(prefix):]
	// orgId is everything up to the next slash (or end of string)
	if slashIdx := strings.Index(rest, "/"); slashIdx >= 0 {
		return rest[:slashIdx]
	}
	return rest
}

func (h *ProxyHandler) CtlAPIDocsProxy(c *gin.Context) {
	proxy := h.reverseProxy(h.cfg.NuonAPIURL + "/docs")
	if proxy == nil {
		c.Status(http.StatusBadGateway)
		return
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *ProxyHandler) APILivez(c *gin.Context) {
	respondJSON(c, http.StatusOK, gin.H{"status": "ok"})
}

func (h *ProxyHandler) AdminCtlAPIProxy(c *gin.Context) {
	proxy := h.reverseProxy(h.cfg.AdminAPIURL)
	if proxy == nil {
		c.Status(http.StatusBadGateway)
		return
	}

	// Strip /api/admin/ctl-api prefix
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = strings.TrimPrefix(originalPath, "/api/admin/ctl-api")
	if c.Request.URL.RawPath != "" {
		c.Request.URL.RawPath = strings.TrimPrefix(c.Request.URL.RawPath, "/api/admin/ctl-api")
	}

	if token, _ := cctx.TokenFromGinContext(c); token != "" {
		c.Request.Header.Set("Authorization", "Bearer "+token)
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
