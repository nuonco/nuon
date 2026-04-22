package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
)

type Params struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	TClient temporalclient.Client
}

type Activities struct {
	db      *gorm.DB
	tClient temporalclient.Client
}

func New(params Params) *Activities {
	return &Activities{
		db:      params.DB,
		tClient: params.TClient,
	}
}
