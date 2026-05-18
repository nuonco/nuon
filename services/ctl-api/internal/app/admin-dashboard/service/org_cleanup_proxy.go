package service

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ProxyDeprovisionOrg proxies the deprovision request to the internal admin API.
func (s *service) ProxyDeprovisionOrg(c *gin.Context) {
	orgID := c.Param("id")
	path := fmt.Sprintf("/v1/orgs/%s/admin-deprovision", orgID)
	s.proxyToInternalAPI(c, "POST", path, c.Request.Body)
}

// ProxyForgetOrg proxies the forget request to the internal admin API.
func (s *service) ProxyForgetOrg(c *gin.Context) {
	orgID := c.Param("id")
	path := fmt.Sprintf("/v1/orgs/%s/admin-forget", orgID)
	s.proxyToInternalAPI(c, "POST", path, c.Request.Body)
}

// ProxyForgetOrgInstalls proxies the forget-installs request to the internal admin API.
func (s *service) ProxyForgetOrgInstalls(c *gin.Context) {
	orgID := c.Param("id")
	path := fmt.Sprintf("/v1/orgs/%s/admin-forget-installs", orgID)
	s.proxyToInternalAPI(c, "POST", path, c.Request.Body)
}

// ProxyForgetInstall proxies the forget request for a single install to the internal admin API.
func (s *service) ProxyForgetInstall(c *gin.Context) {
	installID := c.Param("id")
	path := fmt.Sprintf("/v1/installs/%s/admin-forget", installID)
	s.proxyToInternalAPI(c, "POST", path, c.Request.Body)
}

// ProxyDeprovisionInstall proxies the deprovision request for a single install.
// It looks up the install's org_id and forwards to the public API with the org context header.
func (s *service) ProxyDeprovisionInstall(c *gin.Context) {
	installID := c.Param("id")

	var install app.Install
	if res := s.db.WithContext(c.Request.Context()).
		Select("id", "org_id").
		Where("id = ?", installID).
		First(&install); res.Error != nil {
		s.l.Error("failed to look up install", zap.Error(res.Error), zap.String("install_id", installID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to look up install"})
		return
	}

	targetURL := fmt.Sprintf("http://localhost:%s/v1/installs/%s/deprovision", s.cfg.HTTPPort, installID)
	target, err := url.Parse(targetURL)
	if err != nil {
		s.l.Error("invalid proxy target", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal proxy error"})
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = target
			req.Method = "POST"
			req.Host = target.Host
			if token, _ := c.Cookie("X-Nuon-Auth"); token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			if email := c.GetHeader("X-Nuon-Admin-Email"); email != "" {
				req.Header.Set("X-Nuon-Admin-Email", email)
			}
			req.Header.Set("X-Nuon-Org-ID", install.OrgID)
		},
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}
