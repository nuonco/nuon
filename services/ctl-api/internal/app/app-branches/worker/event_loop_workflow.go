package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/worker/activities"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) getHandlers() map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error {
	return map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated:      AwaitCreated,
		signals.OperationRestart:      AwaitRestart,
		signals.OperationCheckChanges: AwaitCheckChanges,
		signals.OperationExecuteFlow:  AwaitExecuteFlow,
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
