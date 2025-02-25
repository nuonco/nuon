package migrations

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	AcctClient *account.Client
}

type Migrations struct {
	acctClient *account.Client
}

func New(params Params) *Migrations {
	return &Migrations{
		acctClient: params.AcctClient,
	}
}
