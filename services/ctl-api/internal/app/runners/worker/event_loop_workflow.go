package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationRestart: func(ctx workflow.Context, req signals.RequestSignal) error {
			defer w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				RunnerID: req.ID,
			})
			return AwaitRestart(ctx, req)
		},
		signals.OperationDelete: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDelete(ctx, input)
		},
		signals.OperationCreated: func(ctx workflow.Context, req signals.RequestSignal) error {
			defer w.startHealthCheckWorkflow(ctx, HealthCheckRequest{
				RunnerID: req.ID,
			})
			return AwaitCreated(ctx, req)
		},
		signals.OperationProvision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProvision(ctx, input)
		},
		signals.OperationReprovision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitReprovision(ctx, input)
		},
		signals.OperationDeprovision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDeprovision(ctx, input)
		},
		signals.OperationProcessJob: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProcessJob(ctx, input)
		},
		signals.OperationUpdateVersion: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitUpdateVersion(ctx, input)
		},
		signals.OperationGracefulShutdown: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitGracefulShutdown(ctx, input)
		},
		signals.OperationForceShutdown: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitForceShutdown(ctx, input)
		},
		signals.OperationOfflineCheck: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitOfflineCheck(ctx, input)
		},
		signals.OperationFlushOrphanedJobs: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitFlushOrphanedJobs(ctx, input)
		},

		// independent runner
		signals.OperationProvisionServiceAccount: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProvisionServiceAccount(ctx, input)
		},
		signals.OperationReprovisionServiceAccount: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitReprovisionServiceAccount(ctx, input)
		},
		signals.OperationInstallStackVersionRun: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitInstallStackVersionRun(ctx, input)
		},

		// independent runner management mode
		signals.OperationMngVMShutDown: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitMngVMShutdown(ctx, input)
		},
		signals.OperationMngShutDown: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitMngShutdown(ctx, input)
		},
		signals.OperationMngUpdate: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitMngUpdate(ctx, input)
		},
		signals.OperationMngRestart: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitMngRestart(ctx, input)
		},
		signals.OperationMngFetchToken: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitMngFetchToken(ctx, input)
		},
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
