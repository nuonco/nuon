package handlers

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type PostHogProxyHandler struct {
	l      *zap.Logger
	host   string
	client *http.Client
}

func NewPostHogProxyHandler(cfg *internal.Config, l *zap.Logger) *PostHogProxyHandler {
	return &PostHogProxyHandler{
		l:    l,
		host: strings.TrimSuffix(cfg.PostHogHost, "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *PostHogProxyHandler) RegisterRoutes(e *gin.Engine) error {
	e.Any("/ingest/*proxyPath", h.Handle)
	return nil
}

func (h *PostHogProxyHandler) Handle(c *gin.Context) {
	if h.host == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "posthog proxy not configured"})
		return
	}

	upstreamURL := h.host + c.Param("proxyPath")
	if c.Request.URL.RawQuery != "" {
		upstreamURL += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, upstreamURL, c.Request.Body)
	if err != nil {
		h.l.Error("failed to create upstream request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upstream request"})
		return
	}

	for key, values := range c.Request.Header {
		if _, skip := hopByHopRequestHeaders[strings.ToLower(key)]; skip {
			continue
		}
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}

	clientIP := c.ClientIP()
	if existing := req.Header.Get("X-Forwarded-For"); existing != "" {
		req.Header.Set("X-Forwarded-For", clientIP+", "+existing)
	} else {
		req.Header.Set("X-Forwarded-For", clientIP)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		h.l.Error("upstream request failed", zap.Error(err))
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream request failed"})
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		if _, skip := hopByHopResponseHeaders[strings.ToLower(key)]; skip {
			continue
		}
		for _, v := range values {
			c.Header(key, v)
		}
	}

	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
