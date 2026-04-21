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
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
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
	emitteractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/s3payload"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	signalsactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/signals/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"

	// Import all signals to trigger catalog registration
	_ "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog/allsignals"
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

	app     *fxtest.App
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

		// infrastructure
		fx.Provide(log.New),
		fx.Provide(dblog.New),
		fx.Provide(loops.New),
		fx.Provide(github.New),
		fx.Provide(metrics.New),
		fx.Provide(propagator.New),
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

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

		// helpers (needed by workflowactivities)
		fx.Provide(emitterclient.New),
		fx.Provide(vcshelpers.New),
		fx.Provide(appshelpers.New),

		// all activity providers
		fx.Provide(statusactivities.New),
		fx.Provide(job.New),
		fx.Provide(signaldb.NewPayloadConverter),
		fx.Provide(workflowactivities.New),
		fx.Provide(jobactivities.New),
		fx.Provide(signalsactivities.New),
		fx.Provide(emitteractivities.New),
		fx.Provide(signal.NewSignalLifecycleActivities),
		fx.Provide(queueactivities.New),
		fx.Provide(handleractivities.New),
		fx.Provide(queueclient.New),

		// shared activities aggregation (registers all activities)
		fx.Provide(sharedactivities.New),
		fx.Provide(workflows.NewActivities),

		// test dependencies
		fx.Provide(seed.New),
		fx.Provide(flowclient.New),

		// queue + handler workflows
		fx.Provide(queue.NewWorkflows),
		fx.Provide(handler.NewWorkflows),

		// start the test worker
		fx.Provide(worker.AsWorker(New)),

		// invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {})),

		fx.Populate(&e.service),
	)

	e.app.RequireStart()
}

func (e *FlowTestSuite) TearDownSuite() {
	e.app.RequireStop()
}
