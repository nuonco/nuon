package statusactivities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
)

type Params struct {
	fx.In

	DB *gorm.DB `name:"psql"`
	MW metrics.Writer
}

type Activities struct {
	db *gorm.DB
	mw metrics.Writer
}

func New(params Params) *Activities {
	return &Activities{
		db: params.DB,
		mw: params.MW,
	}
}
