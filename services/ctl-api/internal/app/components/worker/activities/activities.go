package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	runnerhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type Params struct {
	fx.In

	DB            *gorm.DB `name:"psql"`
	Helpers       *helpers.Helpers
	AppsHelpers   *appshelpers.Helpers
	RunnerHelpers *runnerhelpers.Helpers
	VCSHelpers    *vcshelpers.Helpers
	AcctClient    *account.Client
	AuthzClient   *authz.Client
	Cfg           *internal.Config
	Features      *features.Features
	SharedActs    *sharedactivities.Activities
}

type Activities struct {
	db             *gorm.DB
	helpers        *helpers.Helpers
	appsHelpers    *appshelpers.Helpers
	runnersHelpers *runnerhelpers.Helpers
	vcsHelpers     *vcshelpers.Helpers
	acctClient     *account.Client
	authzClient    *authz.Client
	cfg            *internal.Config
	features       *features.Features
	sharedActs     *sharedactivities.Activities
}

func New(params Params) *Activities {
	return &Activities{
		cfg:            params.Cfg,
		db:             params.DB,
		helpers:        params.Helpers,
		appsHelpers:    params.AppsHelpers,
		runnersHelpers: params.RunnerHelpers,
		vcsHelpers:     params.VCSHelpers,
		acctClient:     params.AcctClient,
		authzClient:    params.AuthzClient,
		features:       params.Features,
		sharedActs:     params.SharedActs,
	}
}
