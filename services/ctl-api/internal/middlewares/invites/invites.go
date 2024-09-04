package invites

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	L           *zap.Logger
	DB          *gorm.DB `name:"psql"`
	EvClient    eventloop.Client
	AuthzClient *authz.Client
}

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
		if middlewares.IsPublic(ctx) {
			ctx.Next()
			return
		}

		acct, err := middlewares.FromGinContext(ctx)
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

func New(params Params) *middleware {
	return &middleware{
		l:        params.L,
		db:       params.DB,
		evClient: params.EvClient,
		authz:    params.AuthzClient,
	}
}
