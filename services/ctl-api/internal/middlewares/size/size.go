package size

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m *middleware) Name() string {
	return "size"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return LimitRequestBody(m.cfg.MaxRequestSize)
}

func LimitRequestBody(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request == nil || c.Request.Body == nil || limit <= 0 {
			c.Next()
			return
		}
		if c.Request.ContentLength > limit {
			c.AbortWithStatus(http.StatusRequestEntityTooLarge)
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

type Params struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger
}

func New(params Params) *middleware {
	return &middleware{
		l:   params.L,
		cfg: params.Cfg,
	}
}
