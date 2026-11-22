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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	installdelegationdns "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/dns"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/controlplanejob"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/installcomponenthealth"
)

const (
	defaultNamespace string = "installs"
)

type Worker struct {
	worker.Worker
}

// CronWorker polls the isolated install-crons task queue.
type CronWorker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	V            *validator.Validate
	Cfg          *internal.Config
	Tclient      temporalclient.Client
	Wkflows      *Workflows
	Acts         *activities.Activities
	L            *zap.Logger
	Lc           fx.Lifecycle
	Interceptors []interceptor.WorkerInterceptor `group:"interceptors"`

	SharedActs      *workflows.Activities
	SharedWorkflows *workflows.Workflows

	// InstallHealthActs and InstallHealthWorkflows drive the runnerless-install
	// component health workflow on the control-plane-installs task queue built
	// below. They are registered here, not in SharedActs/SharedWorkflows,
	// because they must run only on that queue, not the installs namespace's
	// primary worker.
	InstallHealthActs      *installcomponenthealth.Activities
	InstallHealthWorkflows *installcomponenthealth.Workflows
}

func New(params WorkerParams) (*Worker, error) {
	wkr, err := buildWorker(params, defaultNamespace, pkgworkflows.APITaskQueue, "installs worker")
	if err != nil {
		return nil, err
	}
	return &Worker{wkr}, nil
}

func NewCronWorker(params WorkerParams) (*CronWorker, error) {
	wkr, err := buildWorker(params, pkgworkflows.InstallCronsNamespace, pkgworkflows.InstallCronsTaskQueue, "install crons worker")
	if err != nil {
		return nil, err
	}
	return &CronWorker{wkr}, nil
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
		DisableRegistrationAliasing:            params.Cfg.TemporalDisableRegistrationAliasing,
	})

	// register activities
	wkr.RegisterActivity(params.Acts)
	wkr.RegisterActivity(installdelegationdns.NewActivities(params.V, params.Cfg))

	for _, acts := range params.SharedActs.AllActivities() {
		wkr.RegisterActivity(acts)
	}

	// register workflows
	for _, wkflow := range params.Wkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}

	for _, wkflow := range params.SharedWorkflows.AllWorkflows() {
		wkr.RegisterWorkflow(wkflow)
	}

	// Install jobs dispatched to runners whose status is not-used
	// (use_runner = false) route through job.ExecuteJob to a
	// controlplanejob.ExecuteControlPlaneJob child workflow, which runs as a
	// child of whatever workflow called it and so inherits this namespace. Its
	// RunJob activity is pinned to the control-plane-installs task queue, so a
	// worker polling that queue in this namespace is required or those jobs
	// hang until their execution timeout. See controlplanejob.NewWorker.
	cpWkr := controlplanejob.NewWorker(client, controlplanejob.WorkerConfig{
		MaxConcurrentActivityExecutionSize: params.Cfg.TemporalCPBuildMaxConcurrentActivities,
		MaxConcurrentActivityTaskPollers:   params.Cfg.TemporalMaxConcurrentActivityTaskPollers,
		Interceptors:                       params.Interceptors,
		WorkflowPanicPolicy:                panicPolicy,
		TaskQueue:                          controlplanejob.TaskQueueInstalls,
	}, params.SharedActs.AllActivities()...)

	// InstallComponentHealthWorkflow runs on this same control-plane-installs
	// queue (installcomponenthealth.TaskQueue), started directly by
	// ReconcileUseRunner rather than as a child workflow, so it needs its own
	// workflow registration here alongside its activity.
	cpWkr.RegisterWorkflow(params.InstallHealthWorkflows.InstallComponentHealthWorkflow)
	for _, act := range params.InstallHealthActs.AllActivities() {
		cpWkr.RegisterActivity(act)
	}

	params.Lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting " + logName)
			go func() {
				wkr.Run(worker.InterruptCh())
			}()
			go func() {
				cpWkr.Run(worker.InterruptCh())
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			return nil
		},
	})

	return wkr, nil
}
