package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	componentshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/features"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/terraformcloud"
)

type Params struct {
	fx.In

	V                 *validator.Validate
	DB                *gorm.DB `name:"psql"`
	Components        *protos.Adapter
	AppsHelpers       *appshelpers.Helpers
	ComponentsHelpers *componentshelpers.Helpers
	RunnersHelpers    *runnershelpers.Helpers
	VCSHelpers        *vcshelpers.Helpers
	Helpers           *helpers.Helpers
	EvClient          eventloop.Client
	AcctClient        *account.Client
	AuthzClient       *authz.Client
	Cfg               *internal.Config
	Features          *features.Features

	// NOTE(jm): this is only used as a stop gap until we improve how our repositorys are managed
	LegacyTFCloudOutputs *terraformcloud.OrgsOutputs
}

type Activities struct {
	v                    *validator.Validate
	db                   *gorm.DB
	cfg                  *internal.Config
	components           *protos.Adapter
	appsHelpers          *appshelpers.Helpers
	componentsHelpers    *componentshelpers.Helpers
	runnersHelpers       *runnershelpers.Helpers
	helpers              *helpers.Helpers
	evClient             eventloop.Client
	acctClient           *account.Client
	authzClient          *authz.Client
	vcsHelpers           *vcshelpers.Helpers
	features             *features.Features
	legacyTFCloudOutputs *terraformcloud.OrgsOutputs
}

func New(params Params) *Activities {
	return &Activities{
		db:                   params.DB,
		v:                    params.V,
		cfg:                  params.Cfg,
		components:           params.Components,
		appsHelpers:          params.AppsHelpers,
		runnersHelpers:       params.RunnersHelpers,
		helpers:              params.Helpers,
		evClient:             params.EvClient,
		acctClient:           params.AcctClient,
		authzClient:          params.AuthzClient,
		vcsHelpers:           params.VCSHelpers,
		componentsHelpers:    params.ComponentsHelpers,
		features:             params.Features,
		legacyTFCloudOutputs: params.LegacyTFCloudOutputs,
	}
}
