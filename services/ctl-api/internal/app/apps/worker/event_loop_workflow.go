package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/loop"
)

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitCreated(ctx, input)
		},
		signals.OperationPollDependencies: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitPollDependencies(ctx, input)
		},
		signals.OperationProvision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProvision(ctx, input)
		},
		signals.OperationReprovision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitReprovision(ctx, input)
		},
		signals.OperationUpdateSandbox: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitUpdateSandbox(ctx, input)
		},
		signals.OperationSyncCustomStacks: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitSyncCustomStacks(ctx, input)
		},
		signals.OperationBuildSandbox: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitBuildSandbox(ctx, input)
		},
		signals.OperationDeprovision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDeprovision(ctx, input)
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
