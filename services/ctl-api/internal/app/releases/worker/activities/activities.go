package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	DB       *gorm.DB `name:"psql"`
	EvClient eventloop.Client
}

type Activities struct {
	db       *gorm.DB
	evClient eventloop.Client
}

func New(params Params) (*Activities, error) {
	return &Activities{
		db:       params.DB,
		evClient: params.EvClient,
	}, nil
}
