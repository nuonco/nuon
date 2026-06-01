package awaitrunnerhealthy

import (
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch"
	dblog "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/log"
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

const defaultNamespace = "default"

type TestService struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	V      *validator.Validate
	L      *zap.Logger
	Seed   *seed.Seeder
	Client *client.Client
}

type SignalTestSuite struct {
	suite.Suite

	app     *fxtest.App
	service TestService
}

func TestSignalSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	t.Skip("TODO: signal validation tests require async error propagation from Temporal workflows")
	suite.Run(t, new(SignalTestSuite))
}

func (s *SignalTestSuite) SetupSuite() {
	s.app = fxtest.New(
		s.T(),
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
		fx.Provide(blob.AsBlob(blob.New)),
		fx.Provide(blobstore.NewService),
		fx.Provide(dataconverter.New),
		fx.Provide(temporal.New),
		fx.Provide(validator.New),
		fx.Provide(notifications.New),
		fx.Provide(authz.New),
		fx.Provide(features.New),
		fx.Provide(account.New),
		fx.Provide(analytics.New),
		fx.Provide(analytics.NewTemporal),

		// cloudformation
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
		fx.Provide(worker.AsWorker(worker.New)),

		// invokers
		fx.Invoke(db.DBGroupParam(func([]*gorm.DB) {})),
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {})),

		fx.Populate(&s.service),
	)

	s.app.RequireStart()
}

func (s *SignalTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *SignalTestSuite) TestValidationFailsWithMissingInstallID() {
	ctx := s.service.Seed.EnsureAccount(s.T().Context(), s.T())
	ctx = s.service.Seed.EnsureOrg(ctx, s.T())

	queue, err := s.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(s.T(), err)
	require.Nil(s.T(), s.service.Client.QueueReady(ctx, queue.ID))

	// Enqueue signal with missing InstallID
	_, err = s.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &Signal{
			WorkflowStepID: "step-123",
		},
	})
	require.NotNil(s.T(), err)
	require.Contains(s.T(), err.Error(), "install_id is required")
}

func (s *SignalTestSuite) TestValidationFailsWithMissingWorkflowStepID() {
	ctx := s.service.Seed.EnsureAccount(s.T().Context(), s.T())
	ctx = s.service.Seed.EnsureOrg(ctx, s.T())

	queue, err := s.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(s.T(), err)
	require.Nil(s.T(), s.service.Client.QueueReady(ctx, queue.ID))

	// Enqueue signal with missing WorkflowStepID
	_, err = s.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &Signal{
			InstallID: "install-123",
		},
	})
	require.NotNil(s.T(), err)
	require.Contains(s.T(), err.Error(), "workflow_step_id is required")
}

// Note: Full execution test would require seeding an install with runner
// This would be added once we have proper test fixtures
