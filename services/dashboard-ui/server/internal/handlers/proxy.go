package handlers

import (
	"encoding/json"
	"io"
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
	// Strip /api/ctl-api prefix so ctl-api sees /v1/...
	apiPath := strings.TrimPrefix(c.Request.URL.Path, "/api/ctl-api")
	targetURL := h.cfg.NuonAPIURL + apiPath
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Build the upstream request
	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	// Auth token and org ID come from cookies (set by auth middleware)
	if token, _ := cctx.TokenFromGinContext(c); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if orgID, _ := cctx.OrgIDFromGinContext(c); orgID != "" {
		req.Header.Set("X-Nuon-Org-ID", orgID)
	}
	req.Header.Set("Content-Type", c.GetHeader("Content-Type"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		respondError(c, http.StatusBadGateway, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		respondError(c, http.StatusBadGateway, err)
		return
	}

	// Wrap in TAPIResponse envelope expected by the frontend
	var raw json.RawMessage = body
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.JSON(resp.StatusCode, gin.H{
			"data":    raw,
			"error":   nil,
			"status":  resp.StatusCode,
			"headers": gin.H{},
		})
	} else {
		c.JSON(resp.StatusCode, gin.H{
			"data":    nil,
			"error":   raw,
			"status":  resp.StatusCode,
			"headers": gin.H{},
		})
	}
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
