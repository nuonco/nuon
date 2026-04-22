package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/loop"
)

type OrgEventLoopRequest struct {
	OrgID       string
	SandboxMode bool
}

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		sigs.OperationCreated: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitCreated(ctx, input)
		},
		sigs.OperationProvision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitProvision(ctx, input)
		},
		sigs.OperationReprovision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitReprovision(ctx, input)
		},
		sigs.OperationDeprovision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDeprovision(ctx, input)
		},
		sigs.OperationForceDeprovision: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitForceDeprovision(ctx, input)
		},
		sigs.OperationRestart: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitRestart(ctx, input)
		},
		sigs.OperationRestartRunners: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitRestartRunners(ctx, input)
		},
		sigs.OperationInviteCreated: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitInviteUser(ctx, input)
		},
		sigs.OperationInviteAccepted: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitInviteAccepted(ctx, input)
		},
		sigs.OperationForceDelete: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitForceDelete(ctx, input)
		},
		sigs.OperationDelete: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitDelete(ctx, input)
		},
		sigs.OperationForceSandboxMode: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitForceSandboxMode(ctx, input)
		},
		sigs.OperationEnableFeatureFlags: func(ctx workflow.Context, input signals.RequestSignal) error {
			return AwaitEnableFeatureFlags(ctx, input)
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
