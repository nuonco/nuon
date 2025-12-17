package headers

import (
	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"go.uber.org/zap"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m *middleware) Name() string {
	return "headers"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Nuon-API-Version", m.cfg.GitRef)
		c.Next()
	}
}

func New(writer metrics.Writer, l *zap.Logger, cfg *internal.Config) *middleware {
	return &middleware{
		l:   l,
		cfg: cfg,
	}
}
