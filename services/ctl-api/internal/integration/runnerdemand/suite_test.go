package runnerdemand_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/operatorservice/v1"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/filecache"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/workflows/worker"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionsservice "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/service"
	runnerservice "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/service"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processhealthcheck"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processinit"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processshutdown"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/triggershutdown"
	runneractivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/auth"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/testworker"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/notifications"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter"
	emitteractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/enqueuer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/temporal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/controlplanejob"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type testServices struct {
	fx.In

	DB       *gorm.DB `name:"psql"`
	CHDB     *gorm.DB `name:"ch"`
	L        *zap.Logger
	Seeder   *testseed.Seeder
	Queues   *queueclient.Client
	Flows    *flowclient.Client
	Temporal temporalclient.Client
	Runners  api.Service            `name:"runner-service"`
	Actions  api.Service            `name:"actions-service"`
	Auth     middlewares.Middleware `name:"runner-auth"`
	Accounts *account.Client
	Authz    *authz.Client
	GitHub   *github.Client
}

type isolatedTemporal struct {
	temporalclient.Client
	namespace string
}

func (c *isolatedTemporal) ExecuteWorkflowInNamespace(ctx context.Context, namespace string, opts tclient.StartWorkflowOptions, workflow any, args ...any) (tclient.WorkflowRun, error) {
	if namespace != c.namespace {
		return nil, fmt.Errorf("integration suite refuses to start workflow in namespace %q", namespace)
	}
	return c.Client.ExecuteWorkflowInNamespace(ctx, namespace, opts, workflow, args...)
}

func startSuite(t *testing.T, extra ...fx.Option) (testServices, string) {
	t.Helper()
	namespace := "runner-demand-" + domains.NewWorkflowID()
	var services testServices
	opts := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{CustomValidator: true}),
		fx.Provide(func() (*filecache.FileCache, error) {
			return filecache.New(filecache.Options{Dir: t.TempDir(), MaxCount: 100, MaxBytes: 10 * 1024 * 1024})
		}),
		fx.Provide(temporal.New),
		fx.Decorate(func(tc temporalclient.Client) temporalclient.Client {
			return &isolatedTemporal{Client: tc, namespace: namespace}
		}),
		fx.Decorate(func(_ *enqueuer.Enqueuer, p enqueuer.Params) *enqueuer.Enqueuer {
			return enqueuer.New(p)
		}),
		fx.Decorate(func(cfg *internal.Config) *internal.Config {
			cfg.TemporalNamespace = namespace
			return cfg
		}),
		fx.Supply(fx.Annotate(namespace, fx.ResultTags(`name:"testworker-namespace"`))),
		fx.Provide(analytics.NewTemporal, notifications.New, kafka.New, heartbeater.New),
		fx.Provide(runnerservice.NewRunnerHeartbeatCache, runnerservice.NewRunnerJobWakeRegistry),
		fx.Provide(fx.Annotate(runnerservice.New, fx.As(new(api.Service)), fx.ResultTags(`name:"runner-service"`))),
		fx.Provide(fx.Annotate(actionsservice.New, fx.As(new(api.Service)), fx.ResultTags(`name:"actions-service"`))),
		fx.Provide(fx.Annotate(auth.New, fx.As(new(middlewares.Middleware)), fx.ResultTags(`name:"runner-auth"`))),
		fx.Provide(statusactivities.New, job.New, workflowactivities.New, jobactivities.New),
		fx.Provide(emitteractivities.New, signal.NewSignalLifecycleActivities, queueactivities.New),
		fx.Provide(controlplanejob.NewActivities, handleractivities.New, sharedactivities.New, workflows.NewActivities),
		fx.Provide(queue.NewWorkflows, handler.NewWorkflows),
		fx.Provide(fx.Annotate(runneractivities.New, fx.As(new(any)), fx.ResultTags(`group:"testworker-activities"`))),
		fx.Provide(emitter.NewWorkflows),
		fx.Provide(worker.AsWorker(func(p testworker.WorkerParams, emitters *emitter.Workflows) (*testworker.Worker, error) {
			w, err := testworker.New(p)
			if err != nil {
				return nil, err
			}
			for _, workflow := range emitters.All() {
				w.RegisterWorkflow(workflow)
			}
			return w, nil
		})),
		fx.Invoke(worker.WithWorkers(func([]worker.Worker) {})),
		fx.Populate(&services),
	)
	opts = append(opts, extra...)
	fxApp := fxtest.New(t, opts...)
	require.NoError(t, services.DB.Callback().Create().Before("gorm:before_create").Register("test:runner-namespace", func(tx *gorm.DB) {
		if q, ok := tx.Statement.Dest.(*app.Queue); ok && q.Workflow.Namespace == "runners" {
			q.Workflow.Namespace = namespace
		}
	}))
	fxApp.RequireStart()
	t.Cleanup(func() {
		fxApp.RequireStop()
		require.NoError(t, services.DB.Callback().Create().Remove("test:runner-namespace"))
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := services.Temporal.OperatorService().DeleteNamespace(ctx, &operatorservice.DeleteNamespaceRequest{Namespace: namespace})
		require.NoError(t, err)
		services.Temporal.Close()
	})
	return services, namespace
}

func TestRunnerDemand(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	services, namespace := startSuite(t)
	for name, run := range map[string]func(*testing.T, testServices, string){
		"offline_pickup_and_repeated_resume":  testResume,
		"completion_before_park_unwinds":      testFastCompletion,
		"parallel_manual_retry_and_runner":    testParallelRetryAndRunner,
		"plan_refresh_preserves_instructions": testPlanCredentials,
		"cancel_before_pickup":                testCancelOffline,
		"cancel_after_pickup":                 testCancelExecuting,
		"cancel_queued_job":                   testCancelJob,
		"execution_deadline_starts_at_pickup": testExecutionTimeout,
	} {
		t.Run(name, func(t *testing.T) { run(t, services, namespace) })
	}
}
