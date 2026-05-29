package worker

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"

	// Blank imports to register v2 queue signal types in the catalog.
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/healthcheck"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/webhook_subscription"
)

const TemporalNamespace = "vcs"

type Worker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	Cfg     *internal.Config
	Tclient temporalclient.Client
	Acts    *activities.Activities
	L       *zap.Logger
	Lc      fx.Lifecycle

	SharedActivities *workflows.Activities
	SharedWorkflows  *workflows.Workflows
	Interceptors     []interceptor.WorkerInterceptor `group:"interceptors"`
}

func New(params WorkerParams) (*Worker, error) {
	client, err := params.Tclient.GetNamespaceClient(TemporalNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	panicPolicy := worker.BlockWorkflow
	if params.Cfg.TemporalWorkflowFailurePanic {
		panicPolicy = worker.FailWorkflow
	}

	worker.SetStickyWorkflowCacheSize(params.Cfg.TemporalStickyWorkflowCacheSize)
	wkr := worker.New(client, pkgworkflows.APITaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:     params.Cfg.TemporalMaxConcurrentActivities,
		MaxConcurrentWorkflowTaskExecutionSize: params.Cfg.TemporalMaxConcurrentWorkflowTaskExecutionSize,
		MaxConcurrentActivityTaskPollers:       params.Cfg.TemporalMaxConcurrentActivityTaskPollers,
		MaxConcurrentWorkflowTaskPollers:       params.Cfg.TemporalMaxConcurrentWorkflowTaskPollers,
		Interceptors:                           params.Interceptors,
		WorkflowPanicPolicy:                    panicPolicy,
		DeadlockDetectionTimeout:               params.Cfg.TemporalDeadlockDetectionTimeout,
	})

	// register activities
	wkr.RegisterActivity(params.Acts)
	for _, acts := range params.SharedActivities.AllActivities() {
		wkr.RegisterActivity(acts)
	}

	// register shared workflows (queue, handler, emitter)
	for _, wkflow := range params.SharedWorkflows.AllWorkflows() {
		wkr.RegisterWorkflow(wkflow)
	}

	// register custom vcs workflows (health sweep)
	vcsWorkflows := NewWorkflows(params.Cfg)
	for _, wkflow := range vcsWorkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}

	params.Lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			params.L.Info("starting vcs worker")

			// Idempotently register the native vcs health-sweep schedule (no-ops
			// when native scheduling is disabled).
			if err := workflows.EnsureSweepSchedule(startCtx, params.Tclient, TemporalNamespace,
				"vcs-health-sweep", "*/5 * * * *", "VCSHealthSweep", VCSHealthSweepRequest{}); err != nil {
				params.L.Warn("unable to ensure vcs health sweep schedule", zap.Error(err))
			}

			go func() {
				wkr.Run(worker.InterruptCh())
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			return nil
		},
	})

	return &Worker{wkr}, nil
}
