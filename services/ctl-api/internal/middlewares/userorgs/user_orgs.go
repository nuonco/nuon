package userorgs

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/auth"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/global"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/public"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type middleware struct {
	l  *zap.Logger
	db *gorm.DB
}

func (m middleware) Name() string {
	return "user_orgs"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if public.IsPublic(ctx) {
			ctx.Next()
			return
		}

		user, err := auth.FromContext(ctx)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}

		// if the endpoint is an org protected endpoint, make sure the user has access to this org
		if !global.IsGlobal(ctx) {
			org, err := org.FromContext(ctx)
			if err != nil {
				ctx.Error(err)
				ctx.Abort()
				return
			}

			if err := m.validate(ctx, org.ID, user.Subject); err != nil {
				ctx.Error(err)
				ctx.Abort()
				return
			}
		}

		if err := m.handleInvites(ctx, user.Subject, user.Email); err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func New(l *zap.Logger, db *gorm.DB) *middleware {
	return &middleware{
		l:  l,
		db: db,
	}
}
