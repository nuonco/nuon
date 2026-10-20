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

const ddIntakeURL = "https://browser-intake-us5-datadoghq.com"

// ddIntakePaths are the only intake endpoints the browser SDK posts to.
var ddIntakePaths = map[string]struct{}{
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
	l         *zap.Logger
	intakeURL *url.URL
	client    *http.Client
}

func NewDDProxyHandler(l *zap.Logger) *DDProxyHandler {
	return newDDProxyHandler(l, ddIntakeURL)
}

func newDDProxyHandler(l *zap.Logger, intake string) *DDProxyHandler {
	intakeURL, err := url.Parse(intake)
	if err != nil {
		panic(err)
	}
	return &DDProxyHandler{
		l:         l,
		intakeURL: intakeURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
			// A followed redirect would let the intake response steer this
			// unauthenticated route at cluster-internal hosts.
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

// upstreamURL accepts ddforward only as a relative reference to a known intake
// path. Concatenating it onto the intake host let values like ".evil.com/" or
// "@evil.com/" rewrite the host.
func (h *DDProxyHandler) upstreamURL(ddforward string) (*url.URL, bool) {
	fwd, err := url.Parse(ddforward)
	if err != nil || fwd.Scheme != "" || fwd.Host != "" || fwd.User != nil || fwd.Opaque != "" {
		return nil, false
	}
	if _, ok := ddIntakePaths[fwd.Path]; !ok {
		return nil, false
	}
	return &url.URL{
		Scheme:   h.intakeURL.Scheme,
		Host:     h.intakeURL.Host,
		Path:     fwd.Path,
		RawQuery: fwd.RawQuery,
	}, true
}

func (h *DDProxyHandler) Handle(c *gin.Context) {
	ddforward := c.Query("ddforward")
	if ddforward == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing ddforward parameter"})
		return
	}

	upstream, ok := h.upstreamURL(ddforward)
	if !ok {
		h.l.Warn("rejected ddforward parameter", zap.String("ddforward", ddforward))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ddforward parameter"})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, upstream.String(), c.Request.Body)
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
