package invites

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type Params struct {
	fx.In

	L             *zap.Logger
	DB            *gorm.DB `name:"psql"`
	AccountClient *account.Client
}

type middleware struct {
	l        *zap.Logger
	db       *gorm.DB
	accounts *account.Client
}

func (m middleware) Name() string {
	return "invites"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		acct, err := cctx.AccountFromGinContext(ctx)
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
		accounts: params.AccountClient,
	}
}
