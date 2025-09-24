package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Params struct {
	fx.In

	V  *validator.Validate
	DB *gorm.DB `name:"psql"`
}

type Activities struct {
	v  *validator.Validate
	db *gorm.DB `name:"psql"`
}

func New(params Params) *Activities {
	return &Activities{
		v:  params.V,
		db: params.DB,
	}
}
