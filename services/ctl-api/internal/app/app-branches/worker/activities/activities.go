package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Params struct {
	fx.In

	V *validator.Validate

	DB *gorm.DB `name:"psql"`
}

type Activities struct {
	v  *validator.Validate
	db *gorm.DB
}

func New(params Params) (*Activities, error) {
	return &Activities{
		v:  params.V,
		db: params.DB,
	}, nil
}
