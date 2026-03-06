package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationPollDependencies: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitPollDependencies(ctx, input)
		},
		signals.OperationProvision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProvision(ctx, input)
		},
		signals.OperationCreated: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitCreated(ctx, input)
		},
		signals.OperationQueueBuild: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitQueueBuild(ctx, input)
		},
		signals.OperationConfigCreated: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitQueueBuild(ctx, input)
		},
		signals.OperationBuild: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitBuild(ctx, input)
		},
		signals.OperationDelete: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDelete(ctx, input)
		},
		signals.OperationRestart: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitRestarted(ctx, input)
		},
		signals.OperationUpdateComponentType: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitUpdateComponentType(ctx, input)
		},
	}

	l := loop.Loop[*signals.Signal, signals.RequestSignal]{
		Cfg:              w.cfg,
		V:                w.v,
		MW:               w.mw,
		Handlers:         handlers,
		NewRequestSignal: signals.NewRequestSignal,
		ExistsHook: func(ctx workflow.Context, req eventloop.EventLoopRequest) (bool, error) {
			// TODO(sdboyer) remove the hardcoded response. Proper code is kept in so the import can remain
			// to avoid possibilty of subtle bugs when its enabled.
			_, _ = activities.AwaitCheckExistsByID(ctx, req.ID)
			return true, nil
		},
	}

	return l.Run(ctx, req, pendingSignals)
}
