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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/testworker/seed"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type TestService struct {
	fx.In

	DB   *gorm.DB `name:"psql"`
	V    *validator.Validate
	L    *zap.Logger
	Seed *seed.Seeder

	QueueClient   *queueclient.Client
	EmitterClient *emitterclient.Client
}

type EmitterTestSuite struct {
	suite.Suite

	app *fxtest.App

	service TestService
}

func TestSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(EmitterTestSuite))
}

func (e *EmitterTestSuite) SetupSuite() {
	e.app = fxtest.New(
		e.T(),
		fx.Provide(internal.NewConfig),

		// various dependencies
		fx.Provide(log.New),
		fx.Provide(dblog.New),
		fx.Provide(loops.New),
		fx.Provide(github.New),
		fx.Provide(metrics.New),
		fx.Provide(propagator.New),
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		fx.Provide(gzip.AsGzip(gzip.New)),
		fx.Provide(largepayload.AsLargePayload(largepayload.New)),
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

		// shared activities and workflows
		fx.Provide(statusactivities.New),
		fx.Provide(job.New),
		fx.Provide(signaldb.NewPayloadConverter),

		// clients
		fx.Provide(queueclient.New),
		fx.Provide(emitterclient.New),

		// test dependencies
		fx.Provide(seed.New),

		// queue package dependencies
		fx.Provide(queueactivities.New),
		fx.Provide(queue.NewWorkflows),
		fx.Provide(handler.NewWorkflows),
		fx.Provide(handleractivities.New),

		// emitter package dependencies
		fx.Provide(activities.New),
		fx.Provide(emitter.NewWorkflows),

		// start the test worker
		fx.Provide(worker.AsWorker(New)),

		// invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {
		})),

		fx.Populate(&e.service),
	)

	e.app.RequireStart()
}

func (e *EmitterTestSuite) TearDownSuite() {
	e.app.RequireStop()
}
