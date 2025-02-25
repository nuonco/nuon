package size

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	limits "github.com/gin-contrib/size"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m *middleware) Name() string {
	return "size"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
	return limits.RequestSizeLimiter(m.cfg.MaxRequestSize)
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
