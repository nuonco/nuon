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
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

const (
	defaultNamespace string = "installs"
)

type Worker struct {
	worker.Worker
}

func New(cfg *internal.Config,
	tclient temporalclient.Client,
	wkflows *Workflows,
	acts *activities.Activities,
	l *zap.Logger,
	lc fx.Lifecycle,
) (*Worker, error) {
	client, err := tclient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	wkr := worker.New(client, workflows.APITaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       []interceptor.WorkerInterceptor{},
		WorkflowPanicPolicy:                worker.FailWorkflow,
	})
	wkr.RegisterActivity(acts)
	wkr.RegisterWorkflow(wkflows.EventLoop)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			l.Info("starting installs worker")
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
