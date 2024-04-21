package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
)

func (o *Hooks) InviteCreated(ctx context.Context, orgID, email string) {
	o.sendSignal(ctx, orgID, worker.Signal{
		Operation: worker.OperationInviteCreated,
		Email:     email,
	})
}
