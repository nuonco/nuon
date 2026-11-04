package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

var analyticsForwardedRequestHeaders = map[string]struct{}{
	"accept":       {},
	"content-type": {},
	"user-agent":   {},
}

var analyticsForwardedResponseHeaders = map[string]struct{}{
	"content-type": {},
}

var allowedIngestPathPrefixes = []string{
	"/e",
	"/i",
	"/s",
	"/flags",
	"/static",
	"/array",
}

const (
	maxIngestBodyBytes     = 5 << 20
	ingestConcurrencyLimit = 20
)

type PostHogProxyHandler struct {
	l        *zap.Logger
	hostURL  *url.URL
	client   *http.Client
	inflight chan struct{}
}

func NewPostHogProxyHandler(cfg *internal.Config, l *zap.Logger) *PostHogProxyHandler {
	h := &PostHogProxyHandler{
		l:        l,
		inflight: make(chan struct{}, ingestConcurrencyLimit),
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	if cfg.PostHogKey == "" {
		return h
	}

	hostURL, err := url.Parse(strings.TrimSuffix(cfg.PostHogHost, "/"))
	if err != nil {
		l.Error("invalid posthog host, disabling ingest proxy", zap.Error(err))
		return h
	}
	h.hostURL = hostURL
	return h
}

func (h *PostHogProxyHandler) RegisterRoutes(e *gin.Engine) error {
	e.POST("/ingest/*proxyPath", h.Handle)
	e.GET("/ingest/*proxyPath", h.Handle)
	return nil
}

func sanitizeIngestPath(raw string) (string, bool) {
	for i := 0; i < len(raw); i++ {
		if raw[i] < 0x20 || raw[i] == 0x7f {
			return "", false
		}
	}

	if raw == "" {
		raw = "/"
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	cleaned := path.Clean(raw)
	if cleaned == "." {
		cleaned = "/"
	}
	if raw != "/" && strings.HasSuffix(raw, "/") && !strings.HasSuffix(cleaned, "/") {
		cleaned += "/"
	}

	for _, prefix := range allowedIngestPathPrefixes {
		if cleaned == prefix || strings.HasPrefix(cleaned, prefix+"/") {
			return cleaned, true
		}
	}
	return "", false
}

func redactUpstreamError(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Op + ": " + urlErr.Err.Error()
	}
	return err.Error()
}

func (h *PostHogProxyHandler) Handle(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")

	if h.hostURL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "posthog proxy not configured"})
		return
	}

	cleanedPath, ok := sanitizeIngestPath(c.Param("proxyPath"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ingest path"})
		return
	}

	select {
	case h.inflight <- struct{}{}:
		defer func() { <-h.inflight }()
	default:
		c.Status(http.StatusTooManyRequests)
		return
	}

	if c.Request.ContentLength > maxIngestBodyBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
		return
	}
	body := http.MaxBytesReader(c.Writer, c.Request.Body, maxIngestBodyBytes)
	defer body.Close()

	target := *h.hostURL
	target.Path = h.hostURL.Path + cleanedPath
	target.RawQuery = c.Request.URL.RawQuery

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, target.String(), body)
	if err != nil {
		h.l.Error("failed to create upstream request", zap.String("path", cleanedPath), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upstream request"})
		return
	}
	if c.Request.ContentLength >= 0 {
		req.ContentLength = c.Request.ContentLength
	}

	for key, values := range c.Request.Header {
		if _, ok := analyticsForwardedRequestHeaders[strings.ToLower(key)]; !ok {
			continue
		}
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}
	req.Header.Set("X-Forwarded-For", c.ClientIP())

	resp, err := h.client.Do(req)
	if err != nil {
		h.l.Error("upstream request failed", zap.String("path", cleanedPath), zap.String("err", redactUpstreamError(err)))
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
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		h.l.Error("failed to copy upstream response body", zap.String("path", cleanedPath), zap.Error(err))
		if hijacker, ok := c.Writer.(http.Hijacker); ok {
			if conn, _, hjErr := hijacker.Hijack(); hjErr == nil {
				conn.Close()
			}
		}
	}
}
