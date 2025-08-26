package timeout

import (
	"context"
	"net/http"
	"sync"

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

// Handler returns a Gin middleware that enforces a maximum request duration.
// If a request takes longer than the configured duration, it will be aborted
func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), m.cfg.MaxRequestDuration)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		var once sync.Once

		go func() {
			defer func() {
				once.Do(func() {
					close(finished)
				})
			}()
			c.Next()
		}()

		select {
		case <-finished:
			// Request completed
			return
		case <-ctx.Done():
			// Timeout occurred - only respond if we haven't already
			once.Do(func() {
				c.Header("Connection", "close")
				c.AbortWithStatusJSON(http.StatusInternalServerError, stderr.ErrResponse{
					Error:       "timeout",
					UserError:   true,
					Description: "we were unable to complete this request within time.",
				})
			})
			return
		}
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
