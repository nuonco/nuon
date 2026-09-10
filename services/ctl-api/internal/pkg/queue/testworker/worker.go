package testworker

// This package exposes a testworker which is only used for local testing of these workflows.
// In most common cases, the workflows will be embedded / registered as part of a _different_ worker.
import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/go-playground/validator/v10"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	handleractivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

var defaultNamespace = fmt.Sprintf("ctl-api-queue-testworker-%d", time.Now().UnixNano())

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
	StatusActs     *statusactivities.Activities
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
	wkr.RegisterActivity(params.StatusActs)

	// register workflows
	for _, wkflow := range params.QueueWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}
	for _, wkflow := range params.HandlerWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}

	params.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.L.Info("starting installs worker")
			if err := registerTestNamespace(ctx, params.Tclient); err != nil {
				return err
			}
			return wkr.Start()
		},
		OnStop: func(_ context.Context) error {
			wkr.Stop()
			return nil
		},
	})

	return &Worker{wkr}, nil
}

func registerTestNamespace(ctx context.Context, client temporalclient.Client) error {
	_, err := client.WorkflowService().RegisterNamespace(ctx, &workflowservice.RegisterNamespaceRequest{
		Namespace:                        defaultNamespace,
		WorkflowExecutionRetentionPeriod: durationpb.New(24 * time.Hour),
	})
	if err != nil {
		if _, ok := err.(*serviceerror.NamespaceAlreadyExists); !ok {
			return err
		}
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		_, err = client.WorkflowService().DescribeNamespace(ctx, &workflowservice.DescribeNamespaceRequest{
			Namespace: defaultNamespace,
		})
		if err == nil {
			return nil
		}
		if _, ok := err.(*serviceerror.NamespaceNotFound); !ok {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
