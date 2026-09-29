package service

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) proxyToInternalAPI(c *gin.Context, method, path string, body io.Reader) {
	targetURL := fmt.Sprintf("http://localhost:%s%s", s.cfg.InternalHTTPPort, path)
	target, err := url.Parse(targetURL)
	if err != nil {
		s.l.Error("invalid proxy target", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal proxy error"})
		return
	}

	acct, _ := cctx.AccountFromGinContext(c)

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = target
			req.Method = method
			req.Host = target.Host

			if token, _ := c.Cookie("X-Nuon-Auth"); token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			if acct != nil && acct.Email != "" {
				req.Header.Set("X-Nuon-Admin-Email", acct.Email)
			}
		},
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (s *service) ProxyAddSupportUsers(c *gin.Context) {
	orgID := c.Param("id")
	path := fmt.Sprintf("/v1/orgs/%s/admin-support-users", orgID)
	s.proxyToInternalAPI(c, "POST", path, c.Request.Body)
}

func (s *service) ProxyMigrateQueues(c *gin.Context) {
	orgID := c.Param("id")
	path := fmt.Sprintf("/v1/orgs/%s/admin-migrate-queues", orgID)
	s.proxyToInternalAPI(c, "POST", path, c.Request.Body)
}

func (s *service) ProxySeed(c *gin.Context) {
	s.proxyToInternalAPI(c, "POST", "/v1/general/seed", c.Request.Body)
}
