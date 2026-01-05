package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	AppsHelpers *appshelpers.Helpers
}

type Activities struct {
	db          *gorm.DB
	appsHelpers *appshelpers.Helpers
}

func New(params Params) *Activities {
	return &Activities{
		db:          params.DB,
		appsHelpers: params.AppsHelpers,
	}
}
