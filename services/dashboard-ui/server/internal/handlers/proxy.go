package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type authCacheEntry struct {
	email   string
	expires time.Time
}

type ProxyHandler struct {
	cfg            *internal.Config
	l              *zap.Logger
	codecClient    *http.Client
	codecSemaphore chan struct{}
	authCache      sync.Map // token -> authCacheEntry
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
	publicSwaggerProxy := h.newSwaggerProxy(h.cfg.APIUrl, "/public/swagger", "")
	adminSwaggerProxy := h.newSwaggerProxy(h.cfg.AdminAPIUrl, "/admin/swagger", "/admin")

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

	adminDashboardProxy := h.newAdminDashboardProxy(h.cfg.AdminDashboardUrl)

	nuonOnly.POST("/admin/temporal-codec/decode", h.proxyTemporalCodecDecode)
	nuonOnly.GET("/admin/swagger/*path", gin.WrapH(adminSwaggerProxy))
	nuonOnly.Any("/admin/temporal/*path", gin.WrapH(temporalProxy))
	nuonOnly.GET("/_app/*path", gin.WrapH(temporalProxy))
	nuonOnly.Any("/admin/v1/*path", gin.WrapH(adminAPIProxy))
	nuonOnly.Any("/admin/dashboard/*path", gin.WrapH(adminDashboardProxy))

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
// so the spec loads through the proxy, and rewrites the spec's own
// host/schemes/basePath so Swagger UI's "Execute" requests route back through
// the BFF (apiPrefix) instead of hitting the internal upstream host directly.
func (h *ProxyHandler) newSwaggerProxy(upstreamBase, frontendPrefix, apiPrefix string) *httputil.ReverseProxy {
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
			contentType := resp.Header.Get("Content-Type")
			switch {
			case strings.Contains(contentType, "text/html"):
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				resp.Body.Close()
				rewritten := bytes.ReplaceAll(body, specURLOld, specURLNew)
				return setProxyBody(resp, rewritten)
			case strings.Contains(contentType, "application/json"):
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				resp.Body.Close()
				rewritten, err := rewriteSwaggerSpec(body, apiPrefix)
				if err != nil {
					return err
				}
				return setProxyBody(resp, rewritten)
			default:
				return nil
			}
		},
		ErrorLog: zap.NewStdLog(h.l),
	}
}

func setProxyBody(resp *http.Response, body []byte) error {
	resp.Body = io.NopCloser(bytes.NewReader(body))
	resp.ContentLength = int64(len(body))
	resp.Header.Set("Content-Length", strconv.Itoa(len(body)))
	return nil
}

// rewriteSwaggerSpec points the spec's request target at the BFF proxy so
// Swagger UI's "Execute" stays same-origin. For Swagger 2.0 it drops the
// upstream host/schemes (so the browser page origin is used) and rewrites
// basePath; for OpenAPI 3 it rewrites servers. apiPrefix is the BFF path that
// fronts this API ("/admin" for the admin proxy, "" for the public proxy).
func rewriteSwaggerSpec(body []byte, apiPrefix string) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}

	basePath := apiPrefix
	if basePath == "" {
		basePath = "/"
	}

	if _, isV3 := doc["openapi"]; isV3 {
		doc["servers"] = []map[string]string{{"url": basePath}}
	} else {
		delete(doc, "host")
		delete(doc, "schemes")
		doc["basePath"] = basePath
	}

	return json.Marshal(doc)
}

const adminDashboardPrefix = "/admin/dashboard"

func (h *ProxyHandler) newAdminDashboardProxy(upstreamBase string) *httputil.ReverseProxy {
	target, _ := url.Parse(upstreamBase)
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = strings.TrimPrefix(req.URL.Path, adminDashboardPrefix)
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

			rewritten := bytes.ReplaceAll(body, []byte(`href="/`), []byte(`href="`+adminDashboardPrefix+`/`))
			rewritten = bytes.ReplaceAll(rewritten, []byte(`src="/`), []byte(`src="`+adminDashboardPrefix+`/`))

			rewritten = bytes.Replace(rewritten,
				[]byte(`__ADMIN_CONFIG__={`),
				[]byte(`__ADMIN_CONFIG__={"basePath":"`+adminDashboardPrefix+`",`),
				1)

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

const authCacheTTL = 5 * time.Minute

func (h *ProxyHandler) lookupAuth(token string) (email string, ok bool) {
	if v, found := h.authCache.Load(token); found {
		entry := v.(authCacheEntry)
		if time.Now().Before(entry.expires) {
			return entry.email, true
		}
		h.authCache.Delete(token)
	}
	return "", false
}

func (h *ProxyHandler) verifyAndCache(c *gin.Context, token string) (string, error) {
	if email, ok := h.lookupAuth(token); ok {
		return email, nil
	}
	client, err := nuon.New(nuon.WithURL(h.cfg.APIUrl), nuon.WithAuthToken(token))
	if err != nil {
		return "", err
	}
	me, err := client.GetAuthMe(c.Request.Context())
	if err != nil {
		return "", err
	}
	h.authCache.Store(token, authCacheEntry{
		email:   me.Email,
		expires: time.Now().Add(authCacheTTL),
	})
	return me.Email, nil
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
		if _, verifyErr := h.verifyAndCache(c, token); verifyErr != nil {
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
		email, err := h.verifyAndCache(c, token)
		if err != nil || !strings.HasSuffix(email, "@nuon.co") {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
