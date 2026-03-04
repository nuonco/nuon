package auth

import (
	"net/http"
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

	loginURL := m.cfg.AuthServiceUrl + "/?url=" + m.cfg.AppUrl

	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
			c.Next()
			return
		}

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

		if _, err := client.GetCurrentUser(c.Request.Context()); err != nil {
			m.l.Warn("auth check failed", zap.Error(err))
			c.Redirect(http.StatusFound, loginURL)
			c.Abort()
			return
		}

		c.Next()
	}
}
