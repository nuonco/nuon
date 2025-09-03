package timeout

import (
	"net/http"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m *middleware) Name() string {
	return "timeout"
}

func (m *middleware) onTimeout(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, stderr.ErrResponse{
		Error:       "timeout",
		UserError:   true,
		Description: "we were unable to complete this request within time.",
	})
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}

	return timeout.New(
		timeout.WithTimeout(m.cfg.MaxRequestDuration),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(m.onTimeout),
	)
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
