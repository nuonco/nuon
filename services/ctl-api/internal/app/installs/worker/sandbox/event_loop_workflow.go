package sandbox

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func noop(ctx workflow.Context, sreq signals.RequestSignal) error {
	return nil
}

func (w *Workflows) GetHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationDeprovisionSandboxPlan:      AwaitDeprovisionSandboxPlan,
		signals.OperationDeprovisionSandboxApplyPlan: AwaitDeprovisionSandboxApplyPlan,

		signals.OperationReprovisionSandboxPlan:      AwaitReprovisionSandboxPlan,
		signals.OperationReprovisionSandboxApplyPlan: AwaitReprovisionSandboxApplyPlan,

		signals.OperationProvisionSandboxPlan:      AwaitProvisionSandboxPlan,
		signals.OperationProvisionSandboxApplyPlan: AwaitProvisionSandboxApplyPlan,

		signals.OperationRestart: noop,
		signals.OperationCreated: noop,
	}
}

func (w *Workflows) SandboxEventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := w.GetHandlers()
	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
	}

	return l.Run(ctx, req, pendingSignals)
}
