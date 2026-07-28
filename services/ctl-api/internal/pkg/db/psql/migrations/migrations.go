package migrations

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

type Params struct {
	fx.In

	AcctClient     *account.Client
	RunnersHelpers *runnershelpers.Helpers
	L              *zap.Logger
}

type Migrations struct {
	acctClient     *account.Client
	runnersHelpers *runnershelpers.Helpers
	l              *zap.Logger
}

func New(params Params) *Migrations {
	return &Migrations{
		acctClient:     params.AcctClient,
		runnersHelpers: params.RunnersHelpers,
		l:              params.L,
	}
}
