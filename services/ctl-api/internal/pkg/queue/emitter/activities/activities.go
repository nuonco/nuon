package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
)

type Params struct {
	fx.In

	V       *validator.Validate
	DB      *gorm.DB `name:"psql"`
	TClient temporalclient.Client
	L       *zap.Logger
}

type Activities struct {
	v       *validator.Validate
	db      *gorm.DB
	tClient temporalclient.Client
	l       *zap.Logger
}

func New(params Params) *Activities {
	return &Activities{
		v:       params.V,
		db:      params.DB,
		tClient: params.TClient,
		l:       params.L,
	}
}
