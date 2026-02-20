package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

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
	return func(c *gin.Context) {
		// Skip auth for health endpoints
		path := c.Request.URL.Path
		if path == "/livez" || path == "/readyz" || path == "/version" {
			c.Next()
			return
		}

		// Read token from cookie
		token, err := c.Cookie("X-Nuon-Auth")
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"data":    nil,
				"error":   gin.H{"error": "unauthorized", "description": "missing auth token"},
				"status":  401,
				"headers": gin.H{},
			})
			return
		}

		// Validate token by calling the API
		client, err := nuon.New(
			nuon.WithAuthToken(token),
			nuon.WithURL(m.cfg.NuonAPIURL),
		)
		if err != nil {
			m.l.Error("failed to create validation client", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"data":    nil,
				"error":   gin.H{"error": "internal error", "description": "failed to validate token"},
				"status":  500,
				"headers": gin.H{},
			})
			return
		}

		me, err := client.GetCurrentUser(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"data":    nil,
				"error":   gin.H{"error": "unauthorized", "description": "invalid auth token"},
				"status":  401,
				"headers": gin.H{},
			})
			return
		}

		// Set identity on context
		cctx.SetAccountIDGinContext(c, me.ID)
		cctx.SetTokenGinContext(c, token)
		cctx.SetIsEmployeeGinContext(c, strings.HasSuffix(me.Email, "@nuon.co"))

		// Set org ID from route param if present
		if orgID := c.Param("orgId"); orgID != "" {
			cctx.SetOrgIDGinContext(c, orgID)
		}

		c.Next()
	}
}
