package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated: w.created,
	}

	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
		StartupHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) error {
			w.startMetricsWorkflow(ctx)
			return nil
		},
	}

	return l.Run(ctx, req, pendingSignals)
}
