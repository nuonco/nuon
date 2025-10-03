package worker

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	pkgworkflows "github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	orgiam "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/iam"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows"
)

const (
	defaultNamespace string = "orgs"
)

type Worker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	Cfg              *internal.Config
	TClient          temporalclient.Client
	SharedWkflows    *workflows.Workflows
	SharedActivities *workflows.Activities
	WKflows          *Workflows
	Acts             *activities.Activities
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
		MaxConcurrentActivityExecutionSize: params.Cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       params.Interceptors,
		WorkflowPanicPolicy:                panicPolicy,
	})

	wkr.RegisterActivity(params.Acts)
	for _, acts := range params.SharedActivities.AllActivities() {
		wkr.RegisterActivity(acts)
	}

	// register workflows
	for _, wkflow := range params.WKflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}
	for _, wkflow := range params.SharedWkflows.AllWorkflows() {
		wkr.RegisterWorkflow(wkflow)
	}

	// register infra workflows
	wkr.RegisterActivity(orgiam.NewActivities())

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting orgs worker")
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
