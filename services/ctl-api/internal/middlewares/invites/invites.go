package invites

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	authcontext "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/auth/context"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/public"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type middleware struct {
	l        *zap.Logger
	db       *gorm.DB
	evClient eventloop.Client
	authz    *authz.Client
}

func (m middleware) Name() string {
	return "invites"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if public.IsPublic(ctx) {
			ctx.Next()
			return
		}

		acct, err := authcontext.FromContext(ctx)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}

		if err := m.handleInvites(ctx, acct); err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func New(l *zap.Logger,
	db *gorm.DB,
	evClient eventloop.Client,
	authzClient *authz.Client,
) *middleware {
	return &middleware{
		l:        l,
		db:       db,
		evClient: evClient,
		authz:    authzClient,
	}
}
