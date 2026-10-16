package registryauth

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg *internal.Config
	V   *validator.Validate
	DB  *gorm.DB `name:"psql"`
}

type Issuer struct {
	cfg *internal.Config
	v   *validator.Validate
	db  *gorm.DB
}

type Token struct {
	Username      string
	Password      string
	ServerAddress string
}

func New(params Params) *Issuer {
	return &Issuer{
		cfg: params.Cfg,
		v:   params.V,
		db:  params.DB,
	}
}
