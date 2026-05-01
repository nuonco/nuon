package service

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// proxyToInternalAPI forwards a request to the internal admin API (port 8082).
func (s *service) proxyToInternalAPI(c *gin.Context, method, path string, body io.Reader) {
	targetURL := fmt.Sprintf("http://localhost:%s%s", s.cfg.InternalHTTPPort, path)
	target, err := url.Parse(targetURL)
	if err != nil {
		s.l.Error("invalid proxy target", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal proxy error"})
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = target
			req.Method = method
			req.Host = target.Host

			// Forward the admin auth cookie/headers
			if token, _ := c.Cookie("X-Nuon-Auth"); token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			if email := c.GetHeader("X-Nuon-Admin-Email"); email != "" {
				req.Header.Set("X-Nuon-Admin-Email", email)
			}
		},
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

// ProxyAddSupportUsers proxies the add-support-users request to the admin API.
func (s *service) ProxyAddSupportUsers(c *gin.Context) {
	orgID := c.Param("id")
	path := fmt.Sprintf("/v1/orgs/%s/admin-support-users", orgID)
	s.proxyToInternalAPI(c, "POST", path, c.Request.Body)
}

// ProxyMigrateQueues proxies the migrate-queues request to the admin API.
func (s *service) ProxyMigrateQueues(c *gin.Context) {
	orgID := c.Param("id")
	path := fmt.Sprintf("/v1/orgs/%s/admin-migrate-queues", orgID)
	s.proxyToInternalAPI(c, "POST", path, c.Request.Body)
}
