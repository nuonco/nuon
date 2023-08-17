package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/public"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type middleware struct {
	cfg *internal.Config
	l   *zap.Logger
	db  *gorm.DB
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		m.l.Info("auth middleware")
		if public.IsPublic(ctx) {
			ctx.Next()
			return
		}

		ctx.Next()
	}
}

func (m *middleware) Name() string {
	return "auth"
}

func New(l *zap.Logger, cfg *internal.Config, db *gorm.DB) *middleware {
	return &middleware{
		l:   l,
		cfg: cfg,
		db:  db,
	}
}
