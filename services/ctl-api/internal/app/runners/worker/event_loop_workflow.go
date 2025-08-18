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
			return AwaitRestart(ctx, req)
		},
		signals.OperationDelete: AwaitDelete,
		signals.OperationCreated: func(ctx workflow.Context, req signals.RequestSignal) error {
			defer w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				RunnerID: req.ID,
			})
			return AwaitCreated(ctx, req)
		},
		signals.OperationProvision:         AwaitProvision,
		signals.OperationReprovision:       AwaitReprovision,
		signals.OperationDeprovision:       AwaitDeprovision,
		signals.OperationProcessJob:        AwaitProcessJob,
		signals.OperationUpdateVersion:     AwaitUpdateVersion,
		signals.OperationGracefulShutdown:  AwaitGracefulShutdown,
		signals.OperationForceShutdown:     AwaitForceShutdown,
		signals.OperationOfflineCheck:      AwaitOfflineCheck,
		signals.OperationFlushOrphanedJobs: AwaitFlushOrphanedJobs,

		// independent runner
		signals.OperationProvisionServiceAccount:   AwaitProvisionServiceAccount,
		signals.OperationReprovisionServiceAccount: AwaitReprovisionServiceAccount,
		signals.OperationInstallStackVersionRun:    AwaitInstallStackVersionRun,

		// independent runner management mode
		signals.OperationMngVMShutDown: AwaitMngVMShutdown,
		signals.OperationMngShutDown:   AwaitMngShutdown,
		signals.OperationMngUpdate:     AwaitMngUpdate,
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
