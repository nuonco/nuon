package testworker

// This package exposes a testworker which is only used for local testing of these workflows.
// In most common cases, the workflows will be embedded / registered as part of a _different_ worker.
import (
	"context"
	"fmt"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/handler/activities"
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
	Acts           *activities.Activities
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

	// register activities
	wkr.RegisterActivity(params.Acts)
	wkr.RegisterActivity(params.HandlerActs)

	// register workflows
	for _, wkflow := range params.QueueWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}
	for _, wkflow := range params.HandlerWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}

	params.Lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting installs worker")
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
