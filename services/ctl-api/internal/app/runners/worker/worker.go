package worker

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"

	// Blank imports to register v2 queue signal types in the catalog.
	// The queue handler workflow (registered via SharedWorkflows) deserializes signals by type;
	// importing these packages runs their init() which calls catalog.Register().
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/oninactive"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processhealthcheck"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processinit"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processjob"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processshutdown"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/triggershutdown"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/updatetag"
	runner "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/kuberunner"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
)

type Worker struct {
	worker.Worker
}

// HealthcheckCronWorker polls the isolated runner-healthcheck-crons task queue.
type HealthcheckCronWorker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	V       *validator.Validate
	Cfg     *internal.Config
	Tclient temporalclient.Client
	Wkflows *Workflows
	Acts    *activities.Activities
	L       *zap.Logger
	Lc      fx.Lifecycle

	SharedActivities *workflows.Activities
	SharedWorkflows  *workflows.Workflows
	Interceptors     []interceptor.WorkerInterceptor `group:"interceptors"`
}

func New(params WorkerParams) (*Worker, error) {
	wkr, err := buildWorker(params, "runners", pkgworkflows.APITaskQueue, "runners worker")
	if err != nil {
		return nil, err
	}
	return &Worker{wkr}, nil
}

func NewHealthcheckCronWorker(params WorkerParams) (*HealthcheckCronWorker, error) {
	wkr, err := buildWorker(params, pkgworkflows.RunnerHealthcheckCronsNamespace, pkgworkflows.RunnerHealthcheckCronsTaskQueue, "runner healthcheck cron worker")
	if err != nil {
		return nil, err
	}
	return &HealthcheckCronWorker{wkr}, nil
}

func buildWorker(params WorkerParams, namespace string, taskQueue string, logName string) (worker.Worker, error) {
	client, err := params.Tclient.GetNamespaceClient(namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	panicPolicy := worker.BlockWorkflow
	if params.Cfg.TemporalWorkflowFailurePanic {
		panicPolicy = worker.FailWorkflow
	}

	worker.SetStickyWorkflowCacheSize(params.Cfg.TemporalStickyWorkflowCacheSize)
	wkr := worker.New(client, taskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:     params.Cfg.TemporalMaxConcurrentActivities,
		MaxConcurrentWorkflowTaskExecutionSize: params.Cfg.TemporalMaxConcurrentWorkflowTaskExecutionSize,
		MaxConcurrentActivityTaskPollers:       params.Cfg.TemporalMaxConcurrentActivityTaskPollers,
		MaxConcurrentWorkflowTaskPollers:       params.Cfg.TemporalMaxConcurrentWorkflowTaskPollers,
		StickyScheduleToStartTimeout:           params.Cfg.TemporalStickyScheduleToStartTimeout,
		Interceptors:                           params.Interceptors,
		WorkflowPanicPolicy:                    panicPolicy,
		DeadlockDetectionTimeout:               params.Cfg.TemporalDeadlockDetectionTimeout,
	})

	// register activities
	wkr.RegisterActivity(params.Acts)
	for _, acts := range params.SharedActivities.AllActivities() {
		wkr.RegisterActivity(acts)
	}
	wkr.RegisterActivity(runner.NewActivities(params.V, params.Cfg))

	// register workflows
	for _, wkflow := range params.Wkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}
	for _, wkflow := range params.SharedWorkflows.AllWorkflows() {
		wkr.RegisterWorkflow(wkflow)
	}

	params.Lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting " + logName)
			go func() {
				wkr.Run(worker.InterruptCh())
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			return nil
		},
	})

	return wkr, nil
}
