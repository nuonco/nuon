package handlers

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ddIntakeURL = "https://browser-intake-us5-datadoghq.com"

var hopByHopRequestHeaders = map[string]struct{}{
	"host":                {},
	"content-length":      {},
	"transfer-encoding":   {},
	"connection":          {},
	"keep-alive":          {},
	"proxy-authorization": {},
	"te":                  {},
	"upgrade":             {},
	"expect":              {},
}

var hopByHopResponseHeaders = map[string]struct{}{
	"connection":         {},
	"keep-alive":         {},
	"proxy-authenticate": {},
	"transfer-encoding":  {},
	"upgrade":            {},
	"trailer":            {},
}

type DDProxyHandler struct {
	l      *zap.Logger
	client *http.Client
}

func NewDDProxyHandler(l *zap.Logger) *DDProxyHandler {
	return &DDProxyHandler{
		l: l,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *DDProxyHandler) RegisterRoutes(e *gin.Engine) error {
	e.Any("/api/ddp", h.Handle)
	return nil
}

func (h *DDProxyHandler) Handle(c *gin.Context) {
	ddforward := c.Query("ddforward")
	if ddforward == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing ddforward parameter"})
		return
	}

	upstreamURL := ddIntakeURL + ddforward

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
