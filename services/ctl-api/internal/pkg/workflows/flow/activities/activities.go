package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Params struct {
	fx.In

	DB *gorm.DB `name:"psql"`
}

type Activities struct {
	db *gorm.DB
}

func New(params Params) *Activities {
	return &Activities{
		db: params.DB,
	}
}
