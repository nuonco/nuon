package actions

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) getHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		// TODO(sdboyer) a handful more need to be moved here from the main package, but doing so is too awkward without changing runtime behavior so we deferred in first pass
		signals.OperationExecuteActionWorkflow: AwaitExecuteActionWorkflow,
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
		// StartupHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) error {
		// 	w.handleSyncActionWorkflowTriggers(ctx, signals.RequestSignal{
		// 		Signal: &signals.Signal{
		// 			Type: signals.OperationSyncActionWorkflowTriggers,
		// 		},
		// 		EventLoopRequest: req,
		// 	})
		// 	return nil
		// },
	}

	return l.Run(ctx, req, pendingSignals)
}
