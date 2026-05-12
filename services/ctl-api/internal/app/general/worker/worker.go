package worker

import (
	"context"
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
)

const (
	defaultNamespace string = "general"
)

type Worker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	Cfg             *internal.Config
	Tclient         temporalclient.Client
	Wkflows         *Workflows
	Acts            *activities.Activities
	L               *zap.Logger
	LC              fx.Lifecycle
	Interceptors    []interceptor.WorkerInterceptor `group:"interceptors"`
	SharedActs      *workflows.Activities
	SharedWorkflows *workflows.Workflows
}

func New(params WorkerParams) (*Worker, error) {
	client, err := params.Tclient.GetNamespaceClient(defaultNamespace)
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
	})

	// register activities
	wkr.RegisterActivity(params.Acts)
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

	params.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.L.Info("starting general worker")
			go func() {
				wkr.Run(worker.InterruptCh())
			}()
			startMetricsWorkflowOnBoot(ctx, params.Tclient, params.L)
			return nil
		},
		OnStop: func(_ context.Context) error {
			return nil
		},
	})

	return &Worker{wkr}, nil
}

// startMetricsWorkflowOnBoot ensures the general-metrics cron workflow is running.
// Uses WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING so this is idempotent across
// multiple worker boots.
func startMetricsWorkflowOnBoot(ctx context.Context, tClient temporalclient.Client, l *zap.Logger) {
	opts := tclient.StartWorkflowOptions{
		ID:                       metricsWorkflowName,
		TaskQueue:                pkgworkflows.APITaskQueue,
		CronSchedule:             metricsWorkflowCronTab,
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}

	_, err := tClient.ExecuteWorkflowInNamespace(ctx, defaultNamespace, opts, "Metrics")
	if err != nil {
		l.Warn("unable to start general-metrics workflow", zap.Error(err))
		return
	}
	l.Info("general-metrics workflow started")
}
