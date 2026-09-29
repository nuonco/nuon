package helpers

import (
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

const (
	InstallWorkflowsQueueName = queuenames.InstallWorkflowsQueueName

	InstallSignalsQueueName = queuenames.InstallSignalsQueueName

	InstallApprovalsQueueName = queuenames.InstallApprovalsQueueName

	InstallWorkflowStepGroupsQueueName = queuenames.InstallWorkflowStepGroupsQueueName

	InstallWorkflowStepsQueueName = queuenames.InstallWorkflowStepsQueueName

	InstallStateManagerQueueName = queuenames.InstallStateManagerQueueName

	InstallGenerateStepsQueueName = queuenames.InstallGenerateStepsQueueName

	InstallActionWorkflowsQueueName = queuenames.InstallActionWorkflowsQueueName

	InstallDriftWorkflowsQueueName = queuenames.InstallDriftWorkflowsQueueName

	InstallActionCronSignalsQueueName = queuenames.InstallActionCronSignalsQueueName

	InstallComponentHealthQueueName = queuenames.InstallComponentHealthQueueName

	InstallDriftCronSignalsQueueName = queuenames.InstallDriftCronSignalsQueueName
)

type Params struct {
	fx.In

	V                *validator.Validate
	L                *zap.Logger
	Cfg              *internal.Config
	DB               *gorm.DB `name:"psql"`
	ComponentHelpers *componenthelpers.Helpers
	ActionsHelpers   *actionshelpers.Helpers
	RunbooksHelpers  *runbookshelpers.Helpers
	AppsHelpers      *appshelpers.Helpers
	RunnersHelpers   *runnershelpers.Helpers
	QueueClient      *queueclient.Client
	EmitterClient    *emitterclient.Client
	FeaturesClient   *features.Features
	BlobService      blobstore.Service
	MW               metrics.Writer
	MeterProvider    metric.MeterProvider `optional:"true"`
}

type Helpers struct {
	l                *zap.Logger
	cfg              *internal.Config
	componentHelpers *componenthelpers.Helpers
	runnersHelpers   *runnershelpers.Helpers
	appsHelpers      *appshelpers.Helpers
	actionsHelpers   *actionshelpers.Helpers
	runbooksHelpers  *runbookshelpers.Helpers
	db               *gorm.DB
	queueClient      *queueclient.Client
	emitterClient    *emitterclient.Client
	featuresClient   *features.Features
	blobSvc          blobstore.Service
	mw               metrics.Writer
	stateMetrics     *state.Metrics
}

func New(params Params) *Helpers {
	return &Helpers{
		l:                params.L,
		cfg:              params.Cfg,
		componentHelpers: params.ComponentHelpers,
		runnersHelpers:   params.RunnersHelpers,
		actionsHelpers:   params.ActionsHelpers,
		runbooksHelpers:  params.RunbooksHelpers,
		appsHelpers:      params.AppsHelpers,
		db:               params.DB,
		queueClient:      params.QueueClient,
		emitterClient:    params.EmitterClient,
		featuresClient:   params.FeaturesClient,
		blobSvc:          params.BlobService,
		mw:               params.MW,
		stateMetrics:     state.NewMetrics(params.MeterProvider),
	}
}
