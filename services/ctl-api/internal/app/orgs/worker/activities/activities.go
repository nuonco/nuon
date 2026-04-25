package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type Params struct {
	fx.In

	Cfg             *internal.Config
	DB              *gorm.DB `name:"psql"`
	RunnersHelpers  *runnershelpers.Helpers
	Helpers         *helpers.Helpers
	AppsHelpers     *appshelpers.Helpers
	InstallsHelpers *installshelpers.Helpers
	VCSHelpers      *vcshelpers.Helpers
	EVClient        eventloop.Client
	Features        *features.Features
}

type Activities struct {
	db              *gorm.DB
	evClient        eventloop.Client
	runnersHelpers  *runnershelpers.Helpers
	helpers         *helpers.Helpers
	appsHelpers     *appshelpers.Helpers
	installsHelpers *installshelpers.Helpers
	vcsHelpers      *vcshelpers.Helpers
	features        *features.Features
}

func New(params Params) (*Activities, error) {
	return &Activities{
		db:              params.DB,
		evClient:        params.EVClient,
		runnersHelpers:  params.RunnersHelpers,
		helpers:         params.Helpers,
		appsHelpers:     params.AppsHelpers,
		installsHelpers: params.InstallsHelpers,
		vcsHelpers:      params.VCSHelpers,
		features:        params.Features,
	}, nil
}
