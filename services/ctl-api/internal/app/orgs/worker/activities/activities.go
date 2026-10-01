package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	cloudconnectionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/helpers"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/salesforce"
)

type Params struct {
	fx.In

	Cfg                     *internal.Config
	DB                      *gorm.DB `name:"psql"`
	RunnersHelpers          *runnershelpers.Helpers
	Helpers                 *helpers.Helpers
	AppsHelpers             *appshelpers.Helpers
	ComponentHelpers        *componenthelpers.Helpers
	InstallsHelpers         *installshelpers.Helpers
	VCSHelpers              *vcshelpers.Helpers
	CloudConnectionsHelpers *cloudconnectionshelpers.Helpers
	Features                *features.Features
	Salesforce              salesforce.Client
	AcctClient              *account.Client
}

type Activities struct {
	db                      *gorm.DB
	runnersHelpers          *runnershelpers.Helpers
	helpers                 *helpers.Helpers
	appsHelpers             *appshelpers.Helpers
	componentHelpers        *componenthelpers.Helpers
	installsHelpers         *installshelpers.Helpers
	vcsHelpers              *vcshelpers.Helpers
	cloudConnectionsHelpers *cloudconnectionshelpers.Helpers
	features                *features.Features
	salesforce              salesforce.Client
	acctClient              *account.Client
	enableSupportUsers      bool
}

func New(params Params) (*Activities, error) {
	return &Activities{
		db:                      params.DB,
		runnersHelpers:          params.RunnersHelpers,
		helpers:                 params.Helpers,
		appsHelpers:             params.AppsHelpers,
		componentHelpers:        params.ComponentHelpers,
		installsHelpers:         params.InstallsHelpers,
		vcsHelpers:              params.VCSHelpers,
		cloudConnectionsHelpers: params.CloudConnectionsHelpers,
		features:                params.Features,
		salesforce:              params.Salesforce,
		acctClient:              params.AcctClient,
		enableSupportUsers:      params.Cfg != nil && params.Cfg.EnableSupportUsers,
	}, nil
}
