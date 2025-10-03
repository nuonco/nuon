package worker

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	pkgworkflows "github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	installdelegationdns "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/dns"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows"
)

const (
	defaultNamespace string = "installs"
)

type Worker struct {
	worker.Worker
}

type WorkerParams struct {
	fx.In

	V            *validator.Validate
	Cfg          *internal.Config
	Tclient      temporalclient.Client
	Wkflows      *Workflows
	Acts         *activities.Activities
	L            *zap.Logger
	Lc           fx.Lifecycle
	Interceptors []interceptor.WorkerInterceptor `group:"interceptors"`

	SharedActs      *workflows.Activities
	SharedWorkflows *workflows.Workflows
}

func New(params WorkerParams) (*Worker, error) {
	client, err := params.Tclient.GetNamespaceClient(defaultNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace client: %w", err)
	}

	panicPolicy := worker.BlockWorkflow
	if params.Cfg.TemporalWorkflowFailurePanic {
		panicPolicy = worker.FailWorkflow
	}

	worker.SetStickyWorkflowCacheSize(params.Cfg.TemporalStickyWorkflowCacheSize)
	wkr := worker.New(client, pkgworkflows.APITaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: params.Cfg.TemporalMaxConcurrentActivities,
		Interceptors:                       params.Interceptors,
		WorkflowPanicPolicy:                panicPolicy,
		DisableRegistrationAliasing:        true,
	})

	// register activities
	wkr.RegisterActivity(params.Acts)
	wkr.RegisterActivity(installdelegationdns.NewActivities(params.V, params.Cfg))
	for _, acts := range params.SharedActs.AllActivities() {
		wkr.RegisterActivity(acts)
	}

	// register workflows
	for _, wkflow := range params.Wkflows.All() {
		wkr.RegisterWorkflow(wkflow)
	}

	for _, wkflow := range params.SharedWorkflows.AllWorkflows() {
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
