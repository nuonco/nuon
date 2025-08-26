package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) getHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated:            AwaitCreated,
		signals.OperationUpdated:            AwaitUpdated,
		signals.OperationPollDependencies:   AwaitPollDependencies,
		signals.OperationForget:             AwaitForget,
		signals.OperationExecuteFlow:        AwaitExecuteFlow,
		signals.OperationRerunFlow:          AwaitRerunFlow,
		signals.OperationWorkflowApproveAll: AwaitWorkflowApproveAll,
		signals.OperationRestart: func(ctx workflow.Context, req signals.RequestSignal) error {
			AwaitRestarted(ctx, req)
			w.handleSyncActionWorkflowTriggers(ctx, req)
			return nil
		},
		signals.OperationSyncActionWorkflowTriggers: w.handleSyncActionWorkflowTriggers,
		signals.OperationGenerateState:              AwaitGenerateStateAdmin,

		// NOTE(jm): these should be cross account to the runners namespace
		signals.OperationAwaitRunnerHealthy: w.AwaitRunnerHealthy,
		signals.OperationProvisionRunner:    AwaitProvisionRunner,
		signals.OperationReprovisionRunner:  AwaitReprovisionRunner,

		// NOTE(jm): these should be child loops
		signals.OperationProvisionDNS:   AwaitProvisionDNS,
		signals.OperationDeprovisionDNS: AwaitDeprovisionDNS,
		signals.OperationSyncSecrets:    AwaitSyncSecrets,
	}
}

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := w.getHandlers()
	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
		StartupHook:      w.startup,
		ExistsHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) (bool, error) {
			return activities.AwaitCheckExistsByID(ctx, req.ID)
		},
	}

	return l.Run(ctx, req, pendingSignals)
}
