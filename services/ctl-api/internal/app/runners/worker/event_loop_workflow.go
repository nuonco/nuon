package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationRestart: func(ctx workflow.Context, req signals.RequestSignal) error {
			defer w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				RunnerID: req.ID,
			})
			return w.AwaitRestart(ctx, req)
		},
		signals.OperationDelete: w.AwaitDelete,
		signals.OperationCreated: func(ctx workflow.Context, req signals.RequestSignal) error {
			defer w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				RunnerID: req.ID,
			})
			return w.AwaitCreated(ctx, req)
		},
		signals.OperationProvision:        w.AwaitProvision,
		signals.OperationReprovision:      w.AwaitReprovision,
		signals.OperationDeprovision:      w.AwaitDeprovision,
		signals.OperationProcessJob:       w.AwaitProcessJob,
		signals.OperationUpdateVersion:    w.AwaitUpdateVersion,
		signals.OperationGracefulShutdown: w.AwaitGracefulShutdown,
		signals.OperationForceShutdown:    w.AwaitForceShutdown,
		signals.OperationOfflineCheck:     w.AwaitOfflineCheck,
	}

	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
		StartupHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) error {
			w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				RunnerID: req.ID,
			})
			return nil
		},
		ExistsHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) (bool, error) {
			// TODO(sdboyer) remove the hardcoded response. Proper code is kept in so the import can remain
			// to avoid possibilty of subtle bugs when its enabled.
			_, _ = activities.AwaitCheckExistsByID(ctx, req.ID)
			return true, nil
		},
	}

	return l.Run(ctx, req, pendingSignals)
}
