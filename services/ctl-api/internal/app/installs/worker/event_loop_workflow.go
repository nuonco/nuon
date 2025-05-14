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
		signals.OperationCreated:                  w.AwaitCreated,
		signals.OperationPollDependencies:         w.AwaitPollDependencies,
		signals.OperationForget:                   w.AwaitForget,
		signals.OperationReprovisionRunner:        w.AwaitReprovisionRunner,
		signals.OperationProvisionRunner:          w.AwaitProvisionRunner,
		signals.OperationProvisionDNS:             w.AwaitProvisionDNS,
		signals.OperationDeprovisionDNS:           w.AwaitDeprovisionDNS,
		signals.OperationDeprovisionSandbox:       w.subwfSandbox.AwaitDeprovisionSandbox,
		signals.OperationReprovisionSandbox:       w.subwfSandbox.AwaitReprovisionSandbox,
		signals.OperationProvisionSandbox:         w.subwfSandbox.AwaitProvisionSandbox,
		signals.OperationSyncSecrets:              w.AwaitSyncSecrets,
		signals.OperationExecuteWorkflow:          w.AwaitExecuteWorkflow,
		signals.OperationExecuteActionWorkflow:    w.subwfActions.AwaitExecuteActionWorkflow,
		signals.OperationExecuteDeployComponent:   w.subwfComponents.AwaitExecuteDeployComponent,
		signals.OperationExecuteTeardownComponent: w.subwfComponents.AwaitExecuteTeardownComponent,
		signals.OperationRestart: func(ctx workflow.Context, req signals.RequestSignal) error {
			w.AwaitRestarted(ctx, req)
			w.handleSyncActionWorkflowTriggers(ctx, req)
			return nil
		},
		signals.OperationSyncActionWorkflowTriggers:  w.handleSyncActionWorkflowTriggers,
		signals.OperationGenerateInstallStackVersion: w.subwfStack.AwaitGenerateInstallStackVersion,
		signals.OperationAwaitInstallStackVersionRun: w.subwfStack.AwaitInstallStackVersionRun,
		signals.OperationUpdateInstallStackOutputs:   w.subwfStack.AwaitUpdateInstallStackOutputs,
		signals.OperationAwaitRunnerHealthy:          w.AwaitRunnerHealthy,
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
