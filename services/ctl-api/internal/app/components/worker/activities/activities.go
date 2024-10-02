package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"
	runnerhelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	sharedactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

type Params struct {
	fx.In

	Prt              *protos.Adapter
	DB               *gorm.DB `name:"psql"`
	Helpers          *helpers.Helpers
	EvClient         eventloop.Client
	RunnerHelpers    *runnerhelpers.Helpers
	SharedActivities *sharedactivities.Activities
}

type Activities struct {
	db             *gorm.DB
	protos         *protos.Adapter
	helpers        *helpers.Helpers
	evClient       eventloop.Client
	runnersHelpers *runnerhelpers.Helpers

	*sharedactivities.Activities
}

func New(params Params) *Activities {
	return &Activities{
		db:             params.DB,
		protos:         params.Prt,
		helpers:        params.Helpers,
		evClient:       params.EvClient,
		runnersHelpers: params.RunnerHelpers,
		Activities:     params.SharedActivities,
	}
}
