package testworker

// This package exposes a test worker for integration testing the flow package.
// It registers all activities and workflows needed by the execute-flow and
// execute-workflow-step signals.

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const (
	defaultNamespace string = "default"
)

type Worker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	V              *validator.Validate
	Cfg            *internal.Config
	Tclient        temporalclient.Client
	QueueWkflows   *queue.Workflows
	HandlerWkflows *handler.Workflows
	HandlerActs    *handleractivities.Activities
	QueueActs      *queueactivities.Activities
	StatusActs     *statusactivities.Activities
	WorkflowActs   *workflowactivities.Activities
	SharedActs     *sharedactivities.Activities
	L              *zap.Logger
	Lc             fx.Lifecycle
	Interceptors   []interceptor.WorkerInterceptor `group:"interceptors"`
}

func New(params WorkerParams) (*Worker, error) {
	client, err := params.Tclient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	worker.SetStickyWorkflowCacheSize(params.Cfg.TemporalStickyWorkflowCacheSize)
	wkr := worker.New(client, "api", worker.Options{
		MaxConcurrentActivityExecutionSize: params.Cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       params.Interceptors,
		WorkflowPanicPolicy:                worker.FailWorkflow,
		DisableRegistrationAliasing:        true,
	})

	// Register activities
	wkr.RegisterActivity(params.QueueActs)
	wkr.RegisterActivity(params.HandlerActs)
	wkr.RegisterActivity(params.StatusActs)
	wkr.RegisterActivity(params.WorkflowActs)
	wkr.RegisterActivity(params.SharedActs)

	// Register workflows
	for _, wkflow := range params.QueueWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}
	for _, wkflow := range params.HandlerWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}

	params.Lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting flow test worker")
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
