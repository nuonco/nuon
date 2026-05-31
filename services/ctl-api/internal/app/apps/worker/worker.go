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
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/ecrrepository"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
)

const (
	defaultNamespace string = "apps"
)

type Worker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	Cfg        *internal.Config
	Tclient    temporalclient.Client
	Wkflows    *Workflows
	Acts       *activities.Activities
	BranchActs *branchactivities.Activities

	SharedActs      *workflows.Activities
	SharedWorkflows *workflows.Workflows

	Interceptors []interceptor.WorkerInterceptor `group:"interceptors"`

	L  *zap.Logger
	LC fx.Lifecycle
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
		DeadlockDetectionTimeout:               params.Cfg.TemporalDeadlockDetectionTimeout,
	})

	// register activities
	wkr.RegisterActivity(params.Acts)
	wkr.RegisterActivity(params.BranchActs)
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

	// register nested packages
	wkr.RegisterActivity(ecrrepository.NewActivities(&ecrrepository.ActivitiesParams{
		Cfg: params.Cfg,
	}))

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting apps worker")
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
