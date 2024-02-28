package config

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
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
		ctx.Set(ContextKey, m.cfg)
		ctx.Next()
	}
}

func New(l *zap.Logger, cfg *internal.Config) *middleware {
	return &middleware{
		l:   l,
		cfg: cfg,
	}
}
