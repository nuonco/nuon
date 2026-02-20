package helpers

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V                *validator.Validate
	Cfg              *internal.Config
	DB               *gorm.DB `name:"psql"`
	ComponentHelpers *componenthelpers.Helpers
	ActionsHelpers   *actionshelpers.Helpers
	AppsHelpers      *appshelpers.Helpers
	RunnersHelpers   *runnershelpers.Helpers
	EvClient         eventloop.Client
	QueueClient      *queueclient.Client
}

type Helpers struct {
	cfg              *internal.Config
	componentHelpers *componenthelpers.Helpers
	runnersHelpers   *runnershelpers.Helpers
	appsHelpers      *appshelpers.Helpers
	actionsHelpers   *actionshelpers.Helpers
	db               *gorm.DB
	evClient         eventloop.Client
	queueClient      *queueclient.Client
}

func New(params Params) *Helpers {
	return &Helpers{
		cfg:              params.Cfg,
		componentHelpers: params.ComponentHelpers,
		runnersHelpers:   params.RunnersHelpers,
		actionsHelpers:   params.ActionsHelpers,
		appsHelpers:      params.AppsHelpers,
		db:               params.DB,
		evClient:         params.EvClient,
		queueClient:      params.QueueClient,
	}
}
