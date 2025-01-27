package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	RunnersHelpers *runnershelpers.Helpers
	EVClient       eventloop.Client
}

type Activities struct {
	db             *gorm.DB
	evClient       eventloop.Client
	runnersHelpers *runnershelpers.Helpers
}

func New(params Params) (*Activities, error) {
	return &Activities{
		db:             params.DB,
		evClient:       params.EVClient,
		runnersHelpers: params.RunnersHelpers,
	}, nil
}
