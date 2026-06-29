package migrations

import (
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	AcctClient     *account.Client
	RunnersHelpers *runnershelpers.Helpers
}

type Migrations struct {
	acctClient     *account.Client
	runnersHelpers *runnershelpers.Helpers
}

func New(params Params) *Migrations {
	return &Migrations{
		acctClient:     params.AcctClient,
		runnersHelpers: params.RunnersHelpers,
	}
}
