package apiclient

import (
	"net/http"

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
	return "apiclient"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for health endpoints
		path := c.Request.URL.Path
		if path == "/livez" || path == "/readyz" || path == "/version" {
			c.Next()
			return
		}

		token, err := cctx.TokenFromGinContext(c)
		if err != nil {
			// No token means auth middleware didn't run or skipped.
			c.Next()
			return
		}

		client, err := nuon.New(
			nuon.WithAuthToken(token),
			nuon.WithURL(m.cfg.NuonAPIURL),
		)
		if err != nil {
			m.l.Error("failed to create api client", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"data":    nil,
				"error":   gin.H{"error": "internal error", "description": "failed to create api client"},
				"status":  500,
				"headers": gin.H{},
			})
			return
		}

		// Set org ID if available on context
		if orgID, err := cctx.OrgIDFromGinContext(c); err == nil && orgID != "" {
			client.SetOrgID(orgID)
		}

		cctx.SetAPIClientGinContext(c, client)
		c.Next()
	}
}
