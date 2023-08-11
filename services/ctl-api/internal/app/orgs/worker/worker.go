package worker

import (
	"context"
	"fmt"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	defaultNamespace string = "orgs"
)

type Worker struct {
	worker.Worker
}

func New(cfg *internal.Config,
	tclient temporalclient.Client,
	wkflows *Workflows,
	acts *Activities,
	l *zap.Logger,
	lc fx.Lifecycle) (*Worker, error) {
	client, err := tclient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	wkr := worker.New(client, cfg.TemporalTaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       []interceptor.WorkerInterceptor{},
	})
	wkr.RegisterActivity(acts)
	wkr.RegisterWorkflow(wkflows.ProvisionV2)

	ch := make(chan interface{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			l.Info("starting orgs worker")
			go func() {
				wkr.Run(ch)
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			close(ch)
			return nil
		},
	})

	return &Worker{wkr}, nil
}
