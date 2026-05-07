package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type ProxyHandler struct {
	cfg            *internal.Config
	l              *zap.Logger
	codecClient    *http.Client
	codecSemaphore chan struct{}
}

func NewProxyHandler(cfg *internal.Config, l *zap.Logger) *ProxyHandler {
	return &ProxyHandler{
		cfg: cfg,
		l:   l,
		codecClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		codecSemaphore: make(chan struct{}, 10),
	}
}

func (h *ProxyHandler) RegisterRoutes(e *gin.Engine) error {
	// HTML/asset proxy: strips frontend prefix and adds /docs so the page is
	// served from {upstream}/docs/{path}. ModifyResponse rewrites the embedded
	// absolute spec URL so it routes through the proxy instead of hitting the
	// upstream directly.
	publicSwaggerProxy := h.newSwaggerProxy(h.cfg.APIUrl, "/public/swagger")
	adminSwaggerProxy := h.newSwaggerProxy(h.cfg.AdminAPIUrl, "/admin/swagger")

	temporalProxy := h.newTemporalProxy(h.cfg.TemporalUIUrl)

	e.GET("/public/swagger/*path", gin.WrapH(publicSwaggerProxy))

	authed := e.Group("/", h.requireAuth())
	nuonOnly := authed.Group("/", h.requireNuonEmail())
	adminAPITarget, _ := url.Parse(h.cfg.AdminAPIUrl)
	adminAPIProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			token, _ := req.Cookie(authCookie)
			req.URL.Scheme = adminAPITarget.Scheme
			req.URL.Host = adminAPITarget.Host
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/admin")
			req.Host = adminAPITarget.Host
			req.Header.Del("Cookie")
			req.Header.Del("Accept-Encoding")
			if token != nil && token.Value != "" {
				req.Header.Set("Authorization", "Bearer "+token.Value)
			}
		},
		ErrorLog: zap.NewStdLog(h.l),
	}

	nuonOnly.POST("/admin/temporal-codec/decode", h.proxyTemporalCodecDecode)
	nuonOnly.GET("/admin/swagger/*path", gin.WrapH(adminSwaggerProxy))
	nuonOnly.Any("/admin/temporal/*path", gin.WrapH(temporalProxy))
	nuonOnly.GET("/_app/*path", gin.WrapH(temporalProxy))
	nuonOnly.Any("/admin/v1/*path", gin.WrapH(adminAPIProxy))

	return nil
}

// newProxy builds a reverse proxy that strips stripPrefix and prepends addPrefix.
func (h *ProxyHandler) newProxy(upstreamBase, stripPrefix, addPrefix string) *httputil.ReverseProxy {
	target, _ := url.Parse(upstreamBase)
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			path := strings.TrimPrefix(req.URL.Path, stripPrefix)
			req.URL.Path = addPrefix + path
			req.Host = target.Host
			req.Header.Del("Accept-Encoding")
		},
		ErrorLog: zap.NewStdLog(h.l),
	}
}

// newSwaggerProxy builds a reverse proxy for Swagger UI HTML/assets. It strips
// the frontend prefix and adds /docs so assets are fetched from the upstream
// docs path. ModifyResponse rewrites the embedded absolute spec URL (/oapi/v2)
// to route through the proxy's dedicated spec routes instead.
func (h *ProxyHandler) newSwaggerProxy(upstreamBase, frontendPrefix string) *httputil.ReverseProxy {
	target, _ := url.Parse(upstreamBase)
	specURLOld := []byte("url: '/oapi/v2'")
	specURLNew := []byte("url: '" + frontendPrefix + "/oapi/v2'")
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			path := strings.TrimPrefix(req.URL.Path, frontendPrefix)
			if path == "/oapi/v2" || path == "/oapi/v3" {
				req.URL.Path = path
			} else {
				req.URL.Path = "/docs" + path
			}
			req.Host = target.Host
			req.Header.Del("Accept-Encoding")
		},
		ModifyResponse: func(resp *http.Response) error {
			if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
				return nil
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			resp.Body.Close()
			rewritten := bytes.ReplaceAll(body, specURLOld, specURLNew)
			resp.Body = io.NopCloser(bytes.NewReader(rewritten))
			resp.ContentLength = int64(len(rewritten))
			resp.Header.Set("Content-Length", strconv.Itoa(len(rewritten)))
			return nil
		},
		ErrorLog: zap.NewStdLog(h.l),
	}
}

var cspMetaRe = regexp.MustCompile(`(?i)<meta\s+http-equiv=["']content-security-policy["'][^>]*>`)

func (h *ProxyHandler) newTemporalProxy(upstreamBase string) *httputil.ReverseProxy {
	target, _ := url.Parse(upstreamBase)
	baseOld := []byte(`base: ""`)
	baseNew := []byte(`base: "/admin/temporal"`)
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/admin/temporal")
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
			req.Host = target.Host
			req.Header.Del("Accept-Encoding")
		},
		ModifyResponse: func(resp *http.Response) error {
			if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
				return nil
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			resp.Body.Close()
			rewritten := bytes.ReplaceAll(body, baseOld, baseNew)
			rewritten = cspMetaRe.ReplaceAll(rewritten, nil)
			resp.Body = io.NopCloser(bytes.NewReader(rewritten))
			resp.ContentLength = int64(len(rewritten))
			resp.Header.Set("Content-Length", strconv.Itoa(len(rewritten)))
			return nil
		},
		ErrorLog: zap.NewStdLog(h.l),
	}
}

func (h *ProxyHandler) proxyTemporalCodecDecode(c *gin.Context) {
	select {
	case h.codecSemaphore <- struct{}{}:
		defer func() { <-h.codecSemaphore }()
	default:
		c.Status(http.StatusTooManyRequests)
		return
	}

	target := h.cfg.AdminAPIUrl + "/v1/general/temporal-codec/decode"
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 5<<20))
	if err != nil {
		h.l.Error("failed to read codec decode request body", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		h.l.Error("failed to create codec decode request", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.codecClient.Do(req)
	if err != nil {
		h.l.Error("codec decode upstream request failed", zap.Error(err))
		c.Status(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	limitedBody := io.LimitReader(resp.Body, 10<<20)
	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), limitedBody, nil)
}

func (h *ProxyHandler) requireAuth() gin.HandlerFunc {
	loginURL := h.cfg.AuthServiceUrl + "/?url=" + h.cfg.AppUrl
	return func(c *gin.Context) {
		token, err := c.Cookie(authCookie)
		if err != nil || token == "" {
			c.Redirect(http.StatusFound, loginURL)
			c.Abort()
			return
		}
		client, err := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token))
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		h.l.Debug("requireAuth calling GetAuthMe", zap.String("path", c.Request.URL.Path))
		if _, err := client.GetAuthMe(c.Request.Context()); err != nil {
			c.Redirect(http.StatusFound, loginURL)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *ProxyHandler) requireNuonEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(authCookie)
		client, _ := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token))
		h.l.Debug("requireNuonEmail calling GetAuthMe", zap.String("path", c.Request.URL.Path))
		me, err := client.GetAuthMe(c.Request.Context())
		if err != nil || !strings.HasSuffix(me.Email, "@nuon.co") {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
