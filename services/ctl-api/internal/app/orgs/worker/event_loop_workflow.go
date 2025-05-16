package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
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
		sigs.OperationCreated:            AwaitCreated,
		sigs.OperationProvision:          AwaitProvision,
		sigs.OperationReprovision:        AwaitReprovision,
		sigs.OperationDeprovision:        AwaitDeprovision,
		sigs.OperationForceDeprovision:   AwaitForceDeprovision,
		sigs.OperationRestart:            AwaitRestart,
		sigs.OperationRestartRunners:     AwaitRestartRunners,
		sigs.OperationInviteCreated:      AwaitInviteUser,
		sigs.OperationInviteAccepted:     AwaitInviteAccepted,
		sigs.OperationForceDelete:        AwaitForceDelete,
		sigs.OperationDelete:             AwaitDelete,
		sigs.OperationForceSandboxMode:   AwaitForceSandboxMode,
		sigs.OperationEnableFeatureFlags: AwaitEnableFeatureFlags,
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
