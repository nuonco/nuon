package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/loop"
)

const (
	defaultOrgRegion string = "us-west-2"
)

type OrgEventLoopRequest struct {
	OrgID       string
	SandboxMode bool
}

func (w *Workflows) EventLoop(ctx workflow.Context, req eventloop.EventLoopRequest, pendingSignals []*signals.Signal) error {
	handlers := map[eventloop.SignalType]func(workflow.Context, signals.RequestSignal) error{
		signals.OperationCreated:            AwaitCreated,
		signals.OperationProvision:          AwaitProvision,
		signals.OperationReprovision:        AwaitReprovision,
		signals.OperationDeprovision:        AwaitDeprovision,
		signals.OperationForceDeprovision:   AwaitForceDeprovision,
		signals.OperationRestart:            AwaitRestart,
		signals.OperationRestartRunners:     AwaitRestartRunners,
		signals.OperationInviteCreated:      AwaitInviteUser,
		signals.OperationInviteAccepted:     AwaitInviteAccepted,
		signals.OperationForceDelete:        AwaitForceDelete,
		signals.OperationDelete:             AwaitDelete,
		signals.OperationForceSandboxMode:   AwaitForceSandboxMode,
		signals.OperationEnableFeatureFlags: AwaitEnableFeatureFlags,
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
