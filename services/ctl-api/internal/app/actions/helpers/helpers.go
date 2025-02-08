package helpers

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	V   *validator.Validate
	Cfg *internal.Config
	DB  *gorm.DB `name:"psql"`
}

type Helpers struct {
	cfg *internal.Config
	db  *gorm.DB
}

func New(params Params) *Helpers {
	return &Helpers{
		cfg: params.Cfg,
		db:  params.DB,
	}
}
