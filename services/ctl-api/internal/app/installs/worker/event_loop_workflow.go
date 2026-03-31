package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/actions"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/stack"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) getHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitCreated(ctx, input)
		},
		signals.OperationUpdated: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitUpdated(ctx, input)
		},
		signals.OperationPollDependencies: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitPollDependencies(ctx, input)
		},
		signals.OperationForget: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitForget(ctx, input)
		},
		signals.OperationExecuteFlow: func(ctx workflow.Context, input signals.RequestSignal) error {
			w.ensureComponentLoops(ctx, input)
			return AwaitExecuteFlow(ctx, input)
		},
		signals.OperationRerunFlow: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitRerunFlow(ctx, input)
		},
		signals.OperationWorkflowApproveAll: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitWorkflowApproveAll(ctx, input)
		},
		signals.OperationRestart: func(ctx workflow.Context, req signals.RequestSignal) error {
			AwaitRestarted(ctx, req)
			w.handleSyncActionWorkflowTriggers(ctx, req)
			return nil
		},
		signals.OperationSyncActionWorkflowTriggers: w.handleSyncActionWorkflowTriggers,
		signals.OperationGenerateState: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitGenerateStateAdmin(ctx, input)
		},

		// NOTE(jm): these should be cross account to the runners namespace
		signals.OperationAwaitRunnerHealthy: w.AwaitRunnerHealthy,
		signals.OperationProvisionRunner: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProvisionRunner(ctx, input)
		},
		signals.OperationReprovisionRunner: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitReprovisionRunner(ctx, input)
		},

		// install stack update
		signals.OperationUpdateInstallStackOutputs: func(ctx workflow.Context, input signals.RequestSignal) error {
			return stack.AwaitUpdateInstallStackOutputs(ctx, input)
		},

		// adhoc action runs execute directly in main loop
		signals.OperationActionWorkflowRun: func(ctx workflow.Context, input signals.RequestSignal) error {
			return actions.AwaitExecuteActionWorkflowRun(ctx, input)
		},

		// NOTE(jm): these should be child loops
		signals.OperationProvisionDNS: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProvisionDNS(ctx, input)
		},
		signals.OperationDeprovisionDNS: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDeprovisionDNS(ctx, input)
		},
		signals.OperationSyncSecrets: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitSyncSecrets(ctx, input)
		},
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
