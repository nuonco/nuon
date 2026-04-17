package testworker

import (
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/testworker/seed"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/s3payload"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"

	// Import example package so init() registers fake signals in the catalog
	_ "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/testworker/example"
)

type TestService struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	V           *validator.Validate
	L           *zap.Logger
	Seed        *seed.Seeder
	QueueClient *queueclient.Client
	FlowClient  *flowclient.Client
}

type FlowTestSuite struct {
	suite.Suite

	app *fxtest.App

	service TestService
}

func TestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(FlowTestSuite))
}

func (e *FlowTestSuite) SetupSuite() {
	e.app = fxtest.New(
		e.T(),
		fx.Provide(internal.NewConfig),

		// Core infrastructure
		fx.Provide(log.New),
		fx.Provide(dblog.New),
		fx.Provide(loops.New),
		fx.Provide(github.New),
		fx.Provide(metrics.New),
		fx.Provide(propagator.New),
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		// Blobstore + Temporal
		fx.Provide(blobstore.NewService),
		fx.Provide(gzip.AsGzip(gzip.New)),
		fx.Provide(largepayload.AsLargePayload(largepayload.New)),
		fx.Provide(s3payload.AsS3Payload(s3payload.New)),
		fx.Provide(dataconverter.New),
		fx.Provide(temporal.New),
		fx.Provide(validator.New),
		fx.Provide(notifications.New),
		fx.Provide(eventloop.New),
		fx.Provide(authz.New),
		fx.Provide(features.New),
		fx.Provide(account.New),
		fx.Provide(analytics.New),
		fx.Provide(analytics.NewTemporal),
		fx.Provide(cloudformation.NewTemplates),

		// Signal catalog payload converter
		fx.Provide(signaldb.NewPayloadConverter),

		// Shared activities and workflows
		fx.Provide(statusactivities.New),
		fx.Provide(job.New),

		// Queue package
		fx.Provide(queueactivities.New),
		fx.Provide(queue.NewWorkflows),
		fx.Provide(handler.NewWorkflows),
		fx.Provide(handleractivities.New),
		fx.Provide(queueclient.New),

		// Flow-specific dependencies
		fx.Provide(helpers.New),
		fx.Provide(vcshelpers.New),
		fx.Provide(emitterclient.New),
		fx.Provide(workflowactivities.New),
		fx.Provide(sharedactivities.New),
		fx.Provide(flowclient.New),

		// Test dependencies
		fx.Provide(seed.New),
		fx.Provide(worker.AsWorker(New)),

		// Invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {
		})),

		fx.Populate(&e.service),
	)

	e.app.RequireStart()
}

func (e *FlowTestSuite) TearDownSuite() {
	e.app.RequireStop()
}
