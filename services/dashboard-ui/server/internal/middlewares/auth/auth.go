package auth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const cookieName = "X-Nuon-Auth"

type middleware struct {
	cfg *internal.Config
	l   *zap.Logger
}

func New(cfg *internal.Config, l *zap.Logger) *middleware {
	return &middleware{cfg: cfg, l: l}
}

func (m *middleware) Name() string {
	return "auth"
}

func (m *middleware) Handler() gin.HandlerFunc {
	if m.cfg.AuthServiceUrl == "" {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/assets/") ||
			c.Request.URL.Path == "/login" ||
			strings.HasPrefix(c.Request.URL.Path, "/auth/") {
			c.Next()
			return
		}

		returnURL := m.cfg.AppUrl + c.Request.URL.RequestURI()
		loginURL := m.cfg.AuthServiceUrl + "/?url=" + url.QueryEscape(returnURL)

		token, err := c.Cookie(cookieName)
		if err != nil || token == "" {
			c.Redirect(http.StatusFound, loginURL)
			c.Abort()
			return
		}

		client, err := nuon.New(
			nuon.WithURL(m.cfg.APIUrl),
			nuon.WithAuthToken(token),
		)
		if err != nil {
			m.l.Error("failed to create nuon client", zap.Error(err))
			c.Redirect(http.StatusFound, loginURL)
			c.Abort()
			return
		}

		m.l.Debug("auth middleware calling ValidateToken", zap.String("path", c.Request.URL.Path))
		if err := client.ValidateToken(c.Request.Context()); err != nil {
			m.l.Warn("auth check failed", zap.Error(err))
			c.Redirect(http.StatusFound, loginURL)
			c.Abort()
			return
		}

		c.Next()
	}
}
