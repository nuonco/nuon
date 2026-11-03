package migrations

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

type Params struct {
	fx.In

	AcctClient     *account.Client
	RunnersHelpers *runnershelpers.Helpers
	VCSHelpers     *vcshelpers.Helpers
	L              *zap.Logger
}

type Migrations struct {
	acctClient     *account.Client
	runnersHelpers *runnershelpers.Helpers
	vcsHelpers     *vcshelpers.Helpers
	l              *zap.Logger
}

func New(params Params) *Migrations {
	return &Migrations{
		acctClient:     params.AcctClient,
		runnersHelpers: params.RunnersHelpers,
		vcsHelpers:     params.VCSHelpers,
		l:              params.L,
	}
}
