package helpers

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	actionshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/helpers"
	componenthelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/helpers"

	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
)

type Params struct {
	fx.In

	V                *validator.Validate
	Cfg              *internal.Config
	DB               *gorm.DB `name:"psql"`
	ComponentHelpers *componenthelpers.Helpers
	ActionsHelpers   *actionshelpers.Helpers
	RunnersHelpers   *runnershelpers.Helpers
}

type Helpers struct {
	cfg              *internal.Config
	componentHelpers *componenthelpers.Helpers
	runnersHelpers   *runnershelpers.Helpers
	actionsHelpers   *actionshelpers.Helpers
	db               *gorm.DB
}

func New(params Params) *Helpers {
	return &Helpers{
		cfg:              params.Cfg,
		componentHelpers: params.ComponentHelpers,
		runnersHelpers:   params.RunnersHelpers,
		actionsHelpers:   params.ActionsHelpers,
		db:               params.DB,
	}
}
