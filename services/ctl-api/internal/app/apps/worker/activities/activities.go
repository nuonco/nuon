package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
)

type Params struct {
	fx.In

	V       *validator.Validate
	Helpers *helpers.Helpers
	DB      *gorm.DB `name:"psql"`
}

type Activities struct {
	v       *validator.Validate
	db      *gorm.DB
	helpers *helpers.Helpers
}

func New(params Params) (*Activities, error) {
	return &Activities{
		v:       params.V,
		db:      params.DB,
		helpers: params.Helpers,
	}, nil
}
