package worker

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
)

const (
	defaultNamespace string = "orgs"
)

type Worker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	Cfg     *internal.Config
	TClient temporalclient.Client
	WKflows *Workflows
	Acts    *activities.Activities
	L       *zap.Logger
	LC      fx.Lifecycle
}

func New(params WorkerParams) (*Worker, error) {
	client, err := params.TClient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	wkr := worker.New(client, workflows.APITaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: params.Cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       []interceptor.WorkerInterceptor{},
		WorkflowPanicPolicy:                worker.FailWorkflow,
	})

	wkr.RegisterActivity(params.Acts)

	// register workflows
	wkr.RegisterWorkflow(params.WKflows.EventLoop)
	wkr.RegisterWorkflow(params.WKflows.Created)
	wkr.RegisterWorkflow(params.WKflows.Delete)
	wkr.RegisterWorkflow(params.WKflows.Deprovision)
	wkr.RegisterWorkflow(params.WKflows.ForceDelete)
	wkr.RegisterWorkflow(params.WKflows.ForceDeprovision)
	wkr.RegisterWorkflow(params.WKflows.InviteUser)
	wkr.RegisterWorkflow(params.WKflows.Provision)
	wkr.RegisterWorkflow(params.WKflows.Reprovision)

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
