package testworker

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/filecache"
	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/querycollector"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/loops"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/enqueuer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/testworker/seed"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal/dataconverter/blob"
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

	Client *client.Client
}

type EnqueueTestSuite struct {
	suite.Suite

	app *fxtest.App

	service TestService
}

const integrationEventuallyTimeout = 30 * time.Second

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
		fx.Provide(func() *querycollector.Collector { return querycollector.NewCollector(5000) }),
		fx.Provide(psql.AsPSQL(psql.New)),
		fx.Provide(ch.AsCH(ch.New)),

		fx.Provide(blobstore.NewService),
		fx.Provide(func(cfg *internal.Config, l *zap.Logger) *filecache.FileCache {
			cache, err := filecache.New(filecache.Options{
				Dir: cfg.TemporalBlobCacheDir, MaxCount: cfg.TemporalBlobCacheMaxCount,
				MaxBytes: int64(cfg.TemporalBlobCacheMaxSizeMB) * 1024 * 1024,
			})
			if err != nil {
				l.Warn("failed to create blob cache, caching disabled", zap.Error(err))
				return nil
			}
			return cache
		}),
		fx.Provide(gzip.AsGzip(gzip.New)),
		fx.Provide(largepayload.AsLargePayload(largepayload.New)),
		fx.Provide(blob.AsBlob(blob.New)),
		fx.Provide(dataconverter.New),
		fx.Provide(temporal.New),
		fx.Provide(validator.New),
		fx.Provide(notifications.New),
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
		fx.Provide(enqueuer.New),
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

func (e *EnqueueTestSuite) queueReady(ctx context.Context, queueID string) error {
	var err error
	require.Eventually(e.T(), func() bool {
		err = e.service.Client.QueueReady(ctx, queueID)
		return err == nil
	}, integrationEventuallyTimeout, 100*time.Millisecond, "queue %s did not become ready", queueID)
	return err
}

func (e *EnqueueTestSuite) waitForSignalStatus(ctx context.Context, signalID string, expected app.Status) app.QueueSignal {
	var queueSignal app.QueueSignal
	require.Eventually(e.T(), func() bool {
		queueSignal = app.QueueSignal{}
		err := e.service.DB.WithContext(ctx).Where(app.QueueSignal{ID: signalID}).First(&queueSignal).Error
		return err == nil && queueSignal.Status.Status == expected
	}, integrationEventuallyTimeout, 100*time.Millisecond, "signal %s did not reach status %s", signalID, expected)
	return queueSignal
}

type SampleObj struct {
	Field string `validate:"required"`
}
