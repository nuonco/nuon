package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V                *validator.Validate
	DB               *gorm.DB `name:"psql"`
	Cfg              *internal.Config
	L                *zap.Logger
	AppsHelpers      *appshelpers.Helpers
	ComponentHelpers *componenthelpers.Helpers
	ActionsHelpers   *actionshelpers.Helpers
	InstallsHelpers  *installshelpers.Helpers
	OrgsHelpers      *orgshelpers.Helpers
	EvClient         eventloop.Client
	QueueClient      *queueclient.Client
}

type Activities struct {
	v                *validator.Validate
	db               *gorm.DB
	cfg              *internal.Config
	l                *zap.Logger
	appsHelpers      *appshelpers.Helpers
	componentHelpers *componenthelpers.Helpers
	actionsHelpers   *actionshelpers.Helpers
	installsHelpers  *installshelpers.Helpers
	orgsHelpers      *orgshelpers.Helpers
	evClient         eventloop.Client
	queueClient      *queueclient.Client
}

func New(params Params) (*Activities, error) {
	return &Activities{
		v:                params.V,
		db:               params.DB,
		cfg:              params.Cfg,
		l:                params.L,
		appsHelpers:      params.AppsHelpers,
		componentHelpers: params.ComponentHelpers,
		actionsHelpers:   params.ActionsHelpers,
		installsHelpers:  params.InstallsHelpers,
		orgsHelpers:      params.OrgsHelpers,
		evClient:         params.EvClient,
		queueClient:      params.QueueClient,
	}, nil
}
