package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

type Params struct {
	fx.In

	V          *validator.Validate
	DB         *gorm.DB `name:"psql"`
	Cfg        *internal.Config
	AcctClient *account.Client
}

type Activities struct {
	v          *validator.Validate
	db         *gorm.DB `name:"psql"`
	cfg        *internal.Config
	acctClient *account.Client
}

func New(params Params) *Activities {
	return &Activities{
		v:          params.V,
		db:         params.DB,
		cfg:        params.Cfg,
		acctClient: params.AcctClient,
	}
}
