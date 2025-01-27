package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

type Params struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	Components     *protos.Adapter
	AppsHelpers    *appshelpers.Helpers
	RunnersHelpers *runnershelpers.Helpers
	VCSHelpers     *vcshelpers.Helpers
	Helpers        *helpers.Helpers
	EvClient       eventloop.Client
	AcctClient     *account.Client
	AuthzClient    *authz.Client
	Cfg            *internal.Config
}

type Activities struct {
	db             *gorm.DB
	cfg            *internal.Config
	components     *protos.Adapter
	appsHelpers    *appshelpers.Helpers
	runnersHelpers *runnershelpers.Helpers
	helpers        *helpers.Helpers
	evClient       eventloop.Client
	acctClient     *account.Client
	authzClient    *authz.Client
	vcsHelpers     *vcshelpers.Helpers
}

func New(params Params) *Activities {
	return &Activities{
		db:             params.DB,
		cfg:            params.Cfg,
		components:     params.Components,
		appsHelpers:    params.AppsHelpers,
		runnersHelpers: params.RunnersHelpers,
		helpers:        params.Helpers,
		evClient:       params.EvClient,
		acctClient:     params.AcctClient,
		authzClient:    params.AuthzClient,
		vcsHelpers:     params.VCSHelpers,
	}
}
