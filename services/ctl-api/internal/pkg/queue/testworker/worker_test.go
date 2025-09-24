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

	"github.com/powertoolsdev/mono/pkg/workflows/worker"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/analytics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cloudformation"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/psql"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/features"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/loops"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/handler/activities"
	signaldb "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/testworker/seed"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/dataconverter/gzip"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/temporal/dataconverter/largepayload"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

type TestService struct {
	fx.In

	DB   *gorm.DB `name:"psql"`
	V    *validator.Validate
	L    *zap.Logger
	Seed *seed.Seeder

	Client *client.Client
}

type EnqueueTestSuite struct {
	suite.Suite

	app *fxtest.App

	service TestService
}

func TestSuite(t *testing.T) {
        if os.Getenv("INTEGRATION") != "true" {
                t.Skip("INTEGRATION is not set, skipping")
                return
        }

	suite.Run(t, new(EnqueueTestSuite))
}

func (e *EnqueueTestSuite) SetupSuite() {
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

		// test dependencies
		fx.Provide(seed.New),
		fx.Provide(client.New),

		// start the test worker for testing the queue package
		fx.Provide(activities.New),
		fx.Provide(queue.NewWorkflows),
		fx.Provide(handler.NewWorkflows),
		fx.Provide(handleractivities.New),
		fx.Provide(worker.AsWorker(New)),

		// invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {
		})),

		fx.Populate(&e.service),
	)

	e.app.RequireStart()
}

func (e *EnqueueTestSuite) TearDownSuite() {
	e.app.RequireStop()
}

type SampleObj struct {
	Field string `validate:"required"`
}
