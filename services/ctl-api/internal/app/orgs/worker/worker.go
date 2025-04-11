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

func NewActivityWorker(params WorkerParams) (*Worker, error) {
	client, err := params.TClient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	worker.SetStickyWorkflowCacheSize(params.Cfg.TemporalStickyWorkflowCacheSize)
	wkr := worker.New(client, pkgworkflows.APITaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: params.Cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       params.Interceptors,
		WorkflowPanicPolicy:                worker.FailWorkflow,
	})

	registerActivities(wkr, params.Acts, params.SharedActivities)

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting orgs activity worker")
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

func NewWorkflowWorker(params WorkerParams) (*Worker, error) {
	client, err := params.TClient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	wkr := worker.New(client, pkgworkflows.APITaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: params.Cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       params.Interceptors,
		WorkflowPanicPolicy:                worker.FailWorkflow,
	})

	registerWorkflows(wkr, params.WKflows, params.SharedWkflows)

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting orgs workflow worker")
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

func registerActivities(wkr worker.Worker, acts *activities.Activities, sharedActs *workflows.Activities) {
	wkr.RegisterActivity(acts)
	for _, act := range sharedActs.AllActivities() {
		wkr.RegisterActivity(act)
	}
}

func registerWorkflows(wkr worker.Worker, wkflows *Workflows, sharedWorkflows *workflows.Workflows) {
	for _, wkflow := range wkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}
	for _, wkflow := range sharedWorkflows.AllWorkflows() {
		wkr.RegisterWorkflow(wkflow)
	}
}
