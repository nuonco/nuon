package sandbox

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) GetHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationDeprovisionSandboxPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDeprovisionSandboxPlan(ctx, input)
		},
		signals.OperationDeprovisionSandboxApplyPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDeprovisionSandboxApplyPlan(ctx, input)
		},
		signals.OperationReprovisionSandboxPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitReprovisionSandboxPlan(ctx, input)
		},
		signals.OperationReprovisionSandboxApplyPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitReprovisionSandboxApplyPlan(ctx, input)
		},
		signals.OperationProvisionSandboxPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProvisionSandboxPlan(ctx, input)
		},
		signals.OperationProvisionSandboxApplyPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProvisionSandboxApplyPlan(ctx, input)
		},
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
		StartupHook: func(ctx workflow.Context, elr eventloop.EventLoopRequest) error {
			// Drift detection is now handled by emitters reconciled via app_config_updated signal.
			return nil
		},
	}

	return l.Run(ctx, req, pendingSignals)
}
