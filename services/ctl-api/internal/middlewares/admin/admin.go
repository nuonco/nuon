package admin

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

const (
	emailHeaderKey string = "X-Nuon-Admin-Email"
)

type middleware struct {
	l          *zap.Logger
	acctClient *account.Client
	db         *gorm.DB
}

func (m *middleware) Name() string {
	return "admin"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		if err := m.setAccount(ctx); err != nil {
			ctx.Error(err)
			return
		}

		if err := m.setOrgID(ctx); err != nil {
			ctx.Error(err)
			return
		}

		ctx.Next()
	}
}

type Params struct {
	fx.In

	DB         *gorm.DB `name:"psql"`
	L          *zap.Logger
	AcctClient *account.Client
}

func New(params Params) *middleware {
	return &middleware{
		acctClient: params.AcctClient,
		l:          params.L,
		db:         params.DB,
	}
}
