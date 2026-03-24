package helpers

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

type Params struct {
	fx.In

	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	V              *validator.Validate
	AcctClient     *account.Client
	AuthzClient    *authz.Client
	RunnersHelpers *runnershelpers.Helpers
}

type Helpers struct {
	cfg            *internal.Config
	db             *gorm.DB
	v              *validator.Validate
	acctClient     *account.Client
	authzClient    *authz.Client
	runnersHelpers *runnershelpers.Helpers
}

func New(params Params) *Helpers {
	return &Helpers{
		v:              params.V,
		cfg:            params.Cfg,
		db:             params.DB,
		acctClient:     params.AcctClient,
		authzClient:    params.AuthzClient,
		runnersHelpers: params.RunnersHelpers,
	}
}
