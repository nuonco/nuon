package testworker

import (
	"context"
	"fmt"
	"os"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows"
)

// TESTWORKER_NAMESPACE isolates concurrent suite runs: task queues are namespace-scoped, so runs sharing one steal each other's tasks.
var defaultNamespace = func() string {
	if ns := os.Getenv("TESTWORKER_NAMESPACE"); ns != "" {
		return ns
	}
	return fmt.Sprintf("ctl-api-flow-testworker-%d", os.Getpid())
}()

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
	SharedActs     *workflows.Activities
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

	// Register all shared activities (queue, handler, status, flow, lifecycle, client, etc.)
	for _, acts := range params.SharedActs.AllActivities() {
		wkr.RegisterActivity(acts)
	}

	// Register workflows
	for _, wkflow := range params.QueueWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}
	for _, wkflow := range params.HandlerWkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}

	params.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			params.L.Info("starting flow test worker")
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
		_, err = client.WorkflowService().DescribeNamespace(ctx, &workflowservice.DescribeNamespaceRequest{Namespace: defaultNamespace})
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
