package timeout

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m *middleware) Name() string {
	return "timeout"
}

// Handler returns a Gin middleware that monitors request duration.
// If a request takes longer than the configured duration, it will log an error
// but will NOT cancel the request.
func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a timeout context for monitoring only (don't replace the request context)
		timeoutCtx, cancel := context.WithTimeout(context.Background(), m.cfg.MaxRequestDuration)
		defer cancel()

		finished := make(chan struct{})
		var once sync.Once

		// Start monitoring goroutine
		go func() {
			select {
			case <-finished:
				// Request completed normally
				return
			case <-timeoutCtx.Done():
				cl := cctx.GetLogger(c, m.l)
				cl.Error("request timed out",
					zap.Duration("max_duration", m.cfg.MaxRequestDuration),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method))

				metricsCtx, err := cctx.MetricsContextFromGinContext(c)
				if err != nil {
					cl.Error("no metrics context found")
					return
				}
				metricsCtx.IsTimeout = true
				return
			}
		}()

		// Execute the request handlers
		c.Next()

		// Signal that the request is finished
		once.Do(func() {
			close(finished)
		})
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
