package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var ddIntake = url.URL{Scheme: "https", Host: "browser-intake-us5-datadoghq.com"}

// ddforward is attacker-controlled, so only the browser SDK's intake paths
// are forwarded and the upstream host is never taken from it.
var ddForwardPaths = map[string]struct{}{
	"/api/v2/rum":    {},
	"/api/v2/logs":   {},
	"/api/v2/replay": {},
}

var analyticsForwardedRequestHeaders = map[string]struct{}{
	"accept":       {},
	"content-type": {},
	"user-agent":   {},
}

var analyticsForwardedResponseHeaders = map[string]struct{}{
	"content-type": {},
}

type DDProxyHandler struct {
	l      *zap.Logger
	intake url.URL
	client *http.Client
}

func NewDDProxyHandler(l *zap.Logger) *DDProxyHandler {
	return &DDProxyHandler{
		l:      l,
		intake: ddIntake,
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (h *DDProxyHandler) RegisterRoutes(e *gin.Engine) error {
	e.POST("/api/ddp", h.Handle)
	return nil
}

func (h *DDProxyHandler) upstreamURL(ddforward string) (string, bool) {
	fwd, err := url.Parse(ddforward)
	if err != nil || fwd.Scheme != "" || fwd.Host != "" || fwd.User != nil || fwd.Opaque != "" {
		return "", false
	}
	if _, ok := ddForwardPaths[fwd.Path]; !ok {
		return "", false
	}

	u := h.intake
	u.Path = fwd.Path
	u.RawQuery = fwd.RawQuery
	return u.String(), true
}

func (h *DDProxyHandler) Handle(c *gin.Context) {
	ddforward := c.Query("ddforward")
	if ddforward == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing ddforward parameter"})
		return
	}

	upstreamURL, ok := h.upstreamURL(ddforward)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ddforward parameter"})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, upstreamURL, c.Request.Body)
	if err != nil {
		h.l.Error("failed to create upstream request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upstream request"})
		return
	}

	for key, values := range c.Request.Header {
		if _, ok := analyticsForwardedRequestHeaders[strings.ToLower(key)]; !ok {
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
		if _, ok := analyticsForwardedResponseHeaders[strings.ToLower(key)]; !ok {
			continue
		}
		for _, v := range values {
			c.Header(key, v)
		}
	}

	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
