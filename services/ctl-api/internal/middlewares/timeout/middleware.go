package timeout

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m *middleware) Name() string {
	return "timeout"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), m.cfg.MaxRequestDuration)
		defer cancel()

		finished := make(chan struct{})
		var once sync.Once

		go func() {
			select {
			case <-finished:
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

		c.Next()

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
