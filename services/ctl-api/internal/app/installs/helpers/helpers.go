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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const (
	// InstallWorkflowsQueueName is the queue that orchestrates install workflow execution.
	InstallWorkflowsQueueName = "install-workflows"

	// InstallSignalsQueueName is the queue that handles individual install signal execution.
	InstallSignalsQueueName = "install-signals"

	// InstallWorkflowStepGroupsQueueName is the queue that executes workflow step groups.
	InstallWorkflowStepGroupsQueueName = "install-workflow-step-groups"

	// InstallWorkflowStepsQueueName is the queue that executes individual workflow steps
	// as their own signals (when steps-workflows feature is enabled).
	InstallWorkflowStepsQueueName = "install-workflow-steps"
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
	FeaturesClient   *features.Features
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
	featuresClient   *features.Features
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
		featuresClient:   params.FeaturesClient,
	}
}
