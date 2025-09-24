package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
)

type Params struct {
	fx.In

	V       *validator.Validate
	DB      *gorm.DB `name:"psql"`
	TClient temporalclient.Client
}

type Activities struct {
	v       *validator.Validate
	db      *gorm.DB `name:"psql"`
	tclient temporalclient.Client
}

func New(params Params) *Activities {
	return &Activities{
		v:       params.V,
		db:      params.DB,
		tclient: params.TClient,
	}
}
