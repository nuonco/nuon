package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	sharedactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/activities"
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
	Helpers        *helpers.Helpers
	EvClient       eventloop.Client
	SharedActs     *sharedactivities.Activities
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

	*sharedactivities.Activities
}

func New(params Params) *Activities {
	return &Activities{
		db:             params.DB,
		cfg:            params.Cfg,
		components:     params.Components,
		appsHelpers:    params.AppsHelpers,
		runnersHelpers: params.RunnersHelpers,
		helpers:        params.Helpers,
		Activities:     params.SharedActs,
		evClient:       params.EvClient,
		acctClient:     params.AcctClient,
		authzClient:    params.AuthzClient,
	}
}
