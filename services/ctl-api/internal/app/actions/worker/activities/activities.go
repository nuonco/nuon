package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	appshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	installshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/vcs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	DB *gorm.DB `name:"psql"`

	EvClient        eventloop.Client
	AcctClient      *account.Client
	Cfg             *internal.Config
	RunnersHelpers  *runnershelpers.Helpers
	InstallsHelpers *installshelpers.Helpers
	VCSHelpers      *vcshelpers.Helpers
}

type Activities struct {
	db              *gorm.DB
	cfg             *internal.Config
	appsHelpers     *appshelpers.Helpers
	runnersHelpers  *runnershelpers.Helpers
	installsHelpers *installshelpers.Helpers
	vcsHelpers      *vcshelpers.Helpers
	helpers         *helpers.Helpers
	evClient        eventloop.Client
	acctClient      *account.Client
}

func New(params Params) *Activities {
	return &Activities{
		db:              params.DB,
		cfg:             params.Cfg,
		evClient:        params.EvClient,
		acctClient:      params.AcctClient,
		runnersHelpers:  params.RunnersHelpers,
		installsHelpers: params.InstallsHelpers,
		vcsHelpers:      params.VCSHelpers,
	}
}
