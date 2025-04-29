package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/features"
)

type Params struct {
	fx.In

	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	RunnersHelpers *runnershelpers.Helpers
	Helpers        *helpers.Helpers
	EVClient       eventloop.Client
	Features       *features.Features
}

type Activities struct {
	db             *gorm.DB
	evClient       eventloop.Client
	runnersHelpers *runnershelpers.Helpers
	helpers        *helpers.Helpers
	features       *features.Features
}

func New(params Params) (*Activities, error) {
	return &Activities{
		db:             params.DB,
		evClient:       params.EVClient,
		runnersHelpers: params.RunnersHelpers,
		helpers:        params.Helpers,
		features:       params.Features,
	}, nil
}
