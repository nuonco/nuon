package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
)

func (o *Hooks) ForceDelete(ctx context.Context, orgID string) {
	o.l.Info("sending signal")
	o.sendSignal(ctx, orgID, worker.Signal{
		Operation: worker.OperationForceDelete,
	})
}
