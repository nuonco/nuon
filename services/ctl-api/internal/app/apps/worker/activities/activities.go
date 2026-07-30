package activities

import (
	"github.com/go-playground/validator/v10"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type Params struct {
	fx.In

	V               *validator.Validate
	Helpers         *helpers.Helpers
	DB              *gorm.DB `name:"psql"`
	Cfg             *internal.Config
	RunbooksHelpers *runbookshelpers.Helpers
	FeaturesClient  *features.Features
	TClient         temporalclient.Client
}

type Activities struct {
	v               *validator.Validate
	db              *gorm.DB
	helpers         *helpers.Helpers
	cfg             *internal.Config
	runbooksHelpers *runbookshelpers.Helpers
	featuresClient  *features.Features
	tClient         temporalclient.Client
}

func New(params Params) (*Activities, error) {
	return &Activities{
		v:               params.V,
		db:              params.DB,
		helpers:         params.Helpers,
		cfg:             params.Cfg,
		runbooksHelpers: params.RunbooksHelpers,
		featuresClient:  params.FeaturesClient,
		tClient:         params.TClient,
	}, nil
}
