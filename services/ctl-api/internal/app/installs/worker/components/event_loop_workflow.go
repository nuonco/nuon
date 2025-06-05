package components

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) GetHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationExecuteDeployComponentSyncAndPlan: AwaitExecuteDeployComponentSyncAndPlan,
		signals.OperationExecuteDeployComponentApplyPlan:   AwaitExecuteDeployComponentApplyPlan,

		signals.OperationExecuteDeployComponentSyncImage: AwaitExecuteDeployComponentSyncImage,

		signals.OperationExecuteTeardownComponentSyncAndPlan: AwaitExecuteTeardownComponentSyncAndPlan,
		signals.OperationExecuteTeardownComponentApplyPlan:   AwaitExecuteTeardownComponentApplyPlan,
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
		// StartupHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) error {
		//         w.handleSyncActionWorkflowTriggers(ctx, signals.RequestSignal{
		//                 Signal: &signals.Signal{
		//                         Type: signals.OperationSyncActionWorkflowTriggers,
		//                 },
		//                 EventLoopRequest: req,
		//         })
		//         return nil
		// },
	}

	return l.Run(ctx, req, pendingSignals)
}
