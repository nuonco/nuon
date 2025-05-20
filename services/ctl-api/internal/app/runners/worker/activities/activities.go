package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	Cfg           *internal.Config
	DB            *gorm.DB `name:"psql"`
	CHDB          *gorm.DB `name:"ch"`
	Helpers       *helpers.Helpers
	EVClient      eventloop.Client
	AuthzClient   *authz.Client
	AccountClient *account.Client
}

type Activities struct {
	db          *gorm.DB
	chDB        *gorm.DB
	helpers     *helpers.Helpers
	evClient    eventloop.Client
	authzClient *authz.Client
	acctClient  *account.Client
	cfg         *internal.Config
}

func New(params Params) *Activities {
	return &Activities{
		cfg:         params.Cfg,
		db:          params.DB,
		chDB:        params.CHDB,
		helpers:     params.Helpers,
		evClient:    params.EVClient,
		authzClient: params.AuthzClient,
		acctClient:  params.AccountClient,
	}
}
