package cors

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m *middleware) Name() string {
	return "cors"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods: []string{"PUT", "PATCH", "POST", "GET", "OPTIONS"},
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
			"X-Nuon-Org-ID",
			"X-Nuon-Admin-Email",
			"Origin",

			// for temporal codec server
			"X-Namespace",
			"Referer",
			"Sec-Ch-Ua",
			"Sec-Ch-Ua-Mobile",
			"Sec-Ch-Ua-Platform",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Nuon-Page-Next",
			"X-Nuon-Page-Offset",
			"X-Nuon-Page-Limit",
			"X-Nuon-Page-Number",
			"X-Nuon-Install-Workflow-ID",
			"X-Nuon-API-Next",
			"X-Nuon-API-Version",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func New(writer metrics.Writer, l *zap.Logger, cfg *internal.Config) *middleware {
	return &middleware{
		l: l,
	}
}
