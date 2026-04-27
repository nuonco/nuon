package components

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) GetHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationExecuteDeployComponentSyncAndPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitExecuteDeployComponentSyncAndPlan(ctx, input)
		},
		signals.OperationExecuteDeployComponentApplyPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitExecuteDeployComponentApplyPlan(ctx, input)
		},

		signals.OperationExecuteDeployComponentSyncImage: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitExecuteDeployComponentSyncImage(ctx, input)
		},

		signals.OperationExecuteTeardownComponentSyncAndPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitExecuteTeardownComponentSyncAndPlan(ctx, input)
		},
		signals.OperationExecuteTeardownComponentApplyPlan: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitExecuteTeardownComponentApplyPlan(ctx, input)
		},
	}
}

func (w *Workflows) ComponentEventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
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
