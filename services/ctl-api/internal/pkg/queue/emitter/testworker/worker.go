package testworker

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
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
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
	EmitterWkflows *emitter.Workflows
	QueueWkflows   *queue.Workflows
	HandlerWkflows *handler.Workflows
	EmitterActs    *activities.Activities
	QueueActs      *queueactivities.Activities
	QueueClient    *queueclient.Client
	HandlerActs    *handleractivities.Activities
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

	// Register emitter activities
	wkr.RegisterActivity(params.EmitterActs)

	// Register queue activities (needed by emitter)
	wkr.RegisterActivity(params.QueueActs)
	wkr.RegisterActivity(params.HandlerActs)

	// Register queue client activities (for EnqueueSignal)
	for _, act := range queueclient.ClientActivityFns() {
		wkr.RegisterActivity(act)
	}

	// Register emitter workflows
	for _, wkflow := range params.EmitterWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}

	// Register queue workflows (needed for enqueue operations)
	for _, wkflow := range params.QueueWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}
	for _, wkflow := range params.HandlerWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}

	params.Lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting emitter testworker")
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
