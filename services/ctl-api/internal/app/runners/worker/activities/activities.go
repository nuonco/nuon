package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

type Params struct {
	fx.In

	Cfg           *internal.Config
	DB            *gorm.DB `name:"psql"`
	CHDB          *gorm.DB `name:"ch"`
	Helpers       *helpers.Helpers
	AuthzClient   *authz.Client
	AccountClient *account.Client
}

type Activities struct {
	db          *gorm.DB
	chDB        *gorm.DB
	helpers     *helpers.Helpers
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
		authzClient: params.AuthzClient,
		acctClient:  params.AccountClient,
	}
}
