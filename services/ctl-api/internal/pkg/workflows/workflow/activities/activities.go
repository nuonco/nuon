package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	AppsHelpers *appshelpers.Helpers
	TClient     temporalclient.Client
}

type Activities struct {
	db          *gorm.DB
	chDB        *gorm.DB
	appsHelpers *appshelpers.Helpers
	tClient     temporalclient.Client
}

func New(params Params) *Activities {
	return &Activities{
		db:          params.DB,
		chDB:        params.CHDB,
		appsHelpers: params.AppsHelpers,
		tClient:     params.TClient,
	}
}
