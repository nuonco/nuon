package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
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

	// ctl-api swagger/docs proxy
	e.Any("/api/ctl-api/*path", h.CtlAPIProxy)
	e.Any("/public/swagger/*path", h.CtlAPIProxy)
	e.Any("/public/*path", h.CtlAPIDocsProxy)

	// Admin ctl-api proxy
	e.Any("/api/admin/ctl-api/*path", h.AdminCtlAPIProxy)
	e.Any("/admin/swagger/*path", h.AdminCtlAPIProxy)

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
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *ProxyHandler) CtlAPIDocsProxy(c *gin.Context) {
	proxy := h.reverseProxy(h.cfg.NuonAPIURL + "/docs")
	if proxy == nil {
		c.Status(http.StatusBadGateway)
		return
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *ProxyHandler) AdminCtlAPIProxy(c *gin.Context) {
	proxy := h.reverseProxy(h.cfg.AdminAPIURL)
	if proxy == nil {
		c.Status(http.StatusBadGateway)
		return
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}
