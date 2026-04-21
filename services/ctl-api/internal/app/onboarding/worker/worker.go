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
	onboardingactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
)

const (
	defaultNamespace string = "onboardings"
)

type Worker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	Cfg              *internal.Config
	TClient          temporalclient.Client
	SharedWorkflows  *workflows.Workflows
	SharedActivities *workflows.Activities
	OnboardingActs   *onboardingactivities.Activities
	L                *zap.Logger
	LC               fx.Lifecycle
	Interceptors     []interceptor.WorkerInterceptor `group:"interceptors"`
}

func New(params WorkerParams) (*Worker, error) {
	client, err := params.TClient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	worker.SetStickyWorkflowCacheSize(params.Cfg.TemporalStickyWorkflowCacheSize)
	panicPolicy := worker.BlockWorkflow
	if params.Cfg.TemporalWorkflowFailurePanic {
		panicPolicy = worker.FailWorkflow
	}
	wkr := worker.New(client, pkgworkflows.APITaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:     params.Cfg.TemporalMaxConcurrentActivities,
		MaxConcurrentWorkflowTaskExecutionSize: params.Cfg.TemporalMaxConcurrentWorkflowTaskExecutionSize,
		MaxConcurrentActivityTaskPollers:       params.Cfg.TemporalMaxConcurrentActivityTaskPollers,
		MaxConcurrentWorkflowTaskPollers:       params.Cfg.TemporalMaxConcurrentWorkflowTaskPollers,
		Interceptors:                           params.Interceptors,
		WorkflowPanicPolicy:                    panicPolicy,
	})

	// Register onboarding-specific activities
	wkr.RegisterActivity(params.OnboardingActs)

	// Register shared activities
	for _, acts := range params.SharedActivities.AllActivities() {
		wkr.RegisterActivity(acts)
	}

	// Register shared workflows (Queue, Handler, etc.)
	for _, wkflow := range params.SharedWorkflows.AllWorkflows() {
		wkr.RegisterWorkflow(wkflow)
	}

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting onboardings worker")
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
