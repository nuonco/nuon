package testseed

import (
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Params struct {
	fx.In

	DB *gorm.DB `name:"psql"`
}

type Seeder struct {
	db *gorm.DB
}

func New(params Params) *Seeder {
	return &Seeder{
		db: params.DB,
	}
}
