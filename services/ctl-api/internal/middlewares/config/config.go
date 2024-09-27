package config

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m middleware) Name() string {
	return "config"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		middlewares.SetConfigGinContext(ctx, m.cfg)
		ctx.Next()
	}
}

func New(l *zap.Logger, cfg *internal.Config) *middleware {
	return &middleware{
		l:   l,
		cfg: cfg,
	}
}
