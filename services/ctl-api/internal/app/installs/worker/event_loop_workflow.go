package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/actions"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/components"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/sandbox"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/stack"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) getHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated:                  AwaitCreated,
		signals.OperationPollDependencies:         AwaitPollDependencies,
		signals.OperationForget:                   AwaitForget,
		signals.OperationReprovisionRunner:        AwaitReprovisionRunner,
		signals.OperationProvisionRunner:          AwaitProvisionRunner,
		signals.OperationProvisionDNS:             AwaitProvisionDNS,
		signals.OperationDeprovisionDNS:           AwaitDeprovisionDNS,
		signals.OperationDeprovisionSandbox:       sandbox.AwaitDeprovisionSandbox,
		signals.OperationReprovisionSandbox:       sandbox.AwaitReprovisionSandbox,
		signals.OperationProvisionSandbox:         sandbox.AwaitProvisionSandbox,
		signals.OperationSyncSecrets:              AwaitSyncSecrets,
		signals.OperationExecuteWorkflow:          AwaitExecuteWorkflow,
		signals.OperationExecuteActionWorkflow:    actions.AwaitExecuteActionWorkflow,
		signals.OperationExecuteDeployComponent:   components.AwaitExecuteDeployComponent,
		signals.OperationExecuteTeardownComponent: components.AwaitExecuteTeardownComponent,
		signals.OperationRestart: func(ctx workflow.Context, req signals.RequestSignal) error {
			AwaitRestarted(ctx, req)
			w.handleSyncActionWorkflowTriggers(ctx, req)
			return nil
		},
		signals.OperationSyncActionWorkflowTriggers:  w.handleSyncActionWorkflowTriggers,
		signals.OperationGenerateInstallStackVersion: stack.AwaitGenerateInstallStackVersion,
		signals.OperationAwaitInstallStackVersionRun: stack.AwaitInstallStackVersionRun,
		signals.OperationUpdateInstallStackOutputs:   stack.AwaitUpdateInstallStackOutputs,
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
