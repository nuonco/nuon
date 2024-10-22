package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	Acts           *activities.Activities
	RunnersHelpers *runnershelpers.Helpers
	EVClient       eventloop.Client
}

type Activities struct {
	*activities.Activities

	db             *gorm.DB
	evClient       eventloop.Client
	runnersHelpers *runnershelpers.Helpers
}

func New(params Params) (*Activities, error) {
	return &Activities{
		Activities:     params.Acts,
		db:             params.DB,
		evClient:       params.EVClient,
		runnersHelpers: params.RunnersHelpers,
	}, nil
}
