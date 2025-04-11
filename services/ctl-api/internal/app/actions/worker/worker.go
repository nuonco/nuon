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
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows"
)

const (
	defaultNamespace string = "actions"
)

type WorkerParams struct {
	fx.In

	Cfg             *internal.Config
	Tclient         temporalclient.Client
	Wkflows         *Workflows
	Acts            *activities.Activities
	SharedActs      *workflows.Activities
	SharedWorkflows *workflows.Workflows
	L               *zap.Logger
	Interceptors    []interceptor.WorkerInterceptor `group:"interceptors"`

	LC fx.Lifecycle
}

type Worker struct {
	worker.Worker
}

func NewActivityWorker(params WorkerParams) (*Worker, error) {
	client, err := params.Tclient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	worker.SetStickyWorkflowCacheSize(params.Cfg.TemporalStickyWorkflowCacheSize)
	wkr := worker.New(client, pkgworkflows.APITaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: params.Cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       params.Interceptors,
		WorkflowPanicPolicy:                worker.FailWorkflow,
	})

	registerActivities(wkr, params.Acts, params.SharedActs)

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting actions activity worker")
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
	client, err := params.Tclient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	wkr := worker.New(client, pkgworkflows.APITaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: params.Cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       params.Interceptors,
		WorkflowPanicPolicy:                worker.FailWorkflow,
	})

	registerWorkflows(wkr, params.Wkflows, params.SharedWorkflows)

	params.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			params.L.Info("starting actions workflow worker")
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
