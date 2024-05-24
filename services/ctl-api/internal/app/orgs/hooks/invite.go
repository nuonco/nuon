package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/signals"
)

func (o *Hooks) InviteCreated(ctx context.Context, orgID, inviteID string) {
	o.sendSignal(ctx, orgID, signals.Signal{
		Operation: signals.OperationInviteCreated,
		InviteID:  inviteID,
	})
}

func (o *Hooks) InviteAccepted(ctx context.Context, orgID, inviteID string) {
	o.sendSignal(ctx, orgID, signals.Signal{
		Operation: signals.OperationInviteAccepted,
		InviteID:  inviteID,
	})
}
