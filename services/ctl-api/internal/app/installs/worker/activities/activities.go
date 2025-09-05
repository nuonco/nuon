package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	actionhelper "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	componentshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/features"
)

type Params struct {
	fx.In

	V                 *validator.Validate
	DB                *gorm.DB `name:"psql"`
	AppsHelpers       *appshelpers.Helpers
	ComponentsHelpers *componentshelpers.Helpers
	RunnersHelpers    *runnershelpers.Helpers
	VCSHelpers        *vcshelpers.Helpers
	Helpers           *helpers.Helpers
	ActionHelpers     *actionhelper.Helpers
	EvClient          eventloop.Client
	AcctClient        *account.Client
	AuthzClient       *authz.Client
	Cfg               *internal.Config
	Features          *features.Features
	L                 *zap.Logger
}

type Activities struct {
	v                 *validator.Validate
	db                *gorm.DB
	cfg               *internal.Config
	appsHelpers       *appshelpers.Helpers
	componentsHelpers *componentshelpers.Helpers
	runnersHelpers    *runnershelpers.Helpers
	helpers           *helpers.Helpers
	actionHelpers     *actionhelper.Helpers
	evClient          eventloop.Client
	acctClient        *account.Client
	authzClient       *authz.Client
	vcsHelpers        *vcshelpers.Helpers
	features          *features.Features
	l                 *zap.Logger
}

func New(params Params) *Activities {
	return &Activities{
		db:                params.DB,
		v:                 params.V,
		cfg:               params.Cfg,
		appsHelpers:       params.AppsHelpers,
		runnersHelpers:    params.RunnersHelpers,
		actionHelpers:     params.ActionHelpers,
		helpers:           params.Helpers,
		evClient:          params.EvClient,
		acctClient:        params.AcctClient,
		authzClient:       params.AuthzClient,
		vcsHelpers:        params.VCSHelpers,
		componentsHelpers: params.ComponentsHelpers,
		features:          params.Features,
		l:                 params.L,
	}
}
