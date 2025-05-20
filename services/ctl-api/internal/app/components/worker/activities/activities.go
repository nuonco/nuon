package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	runnerhelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	DB            *gorm.DB `name:"psql"`
	Helpers       *helpers.Helpers
	EvClient      eventloop.Client
	RunnerHelpers *runnerhelpers.Helpers
	VCSHelpers    *vcshelpers.Helpers
	AcctClient    *account.Client
	AuthzClient   *authz.Client
	Cfg           *internal.Config
}

type Activities struct {
	db             *gorm.DB
	helpers        *helpers.Helpers
	evClient       eventloop.Client
	runnersHelpers *runnerhelpers.Helpers
	vcsHelpers     *vcshelpers.Helpers
	acctClient     *account.Client
	authzClient    *authz.Client
	cfg            *internal.Config
}

func New(params Params) *Activities {
	return &Activities{
		cfg:            params.Cfg,
		db:             params.DB,
		helpers:        params.Helpers,
		evClient:       params.EvClient,
		runnersHelpers: params.RunnerHelpers,
		vcsHelpers:     params.VCSHelpers,
		acctClient:     params.AcctClient,
		authzClient:    params.AuthzClient,
	}
}
