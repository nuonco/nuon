package cors

import (
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type middleware struct {
	cfg *internal.Config
	l   *zap.Logger
}

func New(cfg *internal.Config, l *zap.Logger) *middleware {
	return &middleware{cfg: cfg, l: l}
}

func (m *middleware) Name() string {
	return "cors"
}

func (m *middleware) Handler() gin.HandlerFunc {
	allowed := ""
	isLocalApp := false
	if m.cfg != nil {
		allowed = originOf(m.cfg.AppUrl)
		isLocalApp = isLoopbackOrigin(m.cfg.AppUrl)
	}
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if allowed != "" && originOf(origin) == allowed {
				return true
			}
			// Local dev tooling (live-reload proxy, worktrees) runs on other
			// loopback ports the BFF has no way to enumerate.
			return isLocalApp && isLoopbackOrigin(origin)
		},
		AllowMethods: []string{"PUT", "PATCH", "POST", "GET", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
			"X-Nuon-Org-ID",
			"Origin",
			"Accept",
			"Cookie",
		},
		ExposeHeaders:    []string{"Content-Length", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// originOf returns scheme://host for a URL or Origin header. A path is ignored.
// An empty or invalid value returns "" so the caller allows nothing.
func originOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + strings.ToLower(u.Host)
}

func isLoopbackOrigin(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" {
		return false
	}
	switch strings.ToLower(u.Hostname()) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
