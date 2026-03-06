package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/actions/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitCreated(ctx, input)
		},
		signals.OperationRestart: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitRestart(ctx, input)
		},
		signals.OperationDelete: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDelete(ctx, input)
		},
		signals.OperationPollDependencies: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitPollDependencies(ctx, input)
		},
		signals.OperationConfigCreated: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitConfigCreated(ctx, input)
		},
	}

	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
	}

	return l.Run(ctx, req, pendingSignals)
}
