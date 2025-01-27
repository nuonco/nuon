package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	runnerhelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

type Params struct {
	fx.In

	Prt           *protos.Adapter
	DB            *gorm.DB `name:"psql"`
	Helpers       *helpers.Helpers
	EvClient      eventloop.Client
	RunnerHelpers *runnerhelpers.Helpers
	AcctClient    *account.Client
	AuthzClient   *authz.Client
	Cfg           *internal.Config
}

type Activities struct {
	db             *gorm.DB
	protos         *protos.Adapter
	helpers        *helpers.Helpers
	evClient       eventloop.Client
	runnersHelpers *runnerhelpers.Helpers
	acctClient     *account.Client
	authzClient    *authz.Client
	cfg            *internal.Config
}

func New(params Params) *Activities {
	return &Activities{
		cfg:            params.Cfg,
		db:             params.DB,
		protos:         params.Prt,
		helpers:        params.Helpers,
		evClient:       params.EvClient,
		runnersHelpers: params.RunnerHelpers,
		acctClient:     params.AcctClient,
		authzClient:    params.AuthzClient,
	}
}
