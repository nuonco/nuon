package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/signals"
)

func (o *Hooks) Deleted(ctx context.Context, orgID string) {
	o.l.Info("sending signal")
	o.sendSignal(ctx, orgID, signals.Signal{
		Operation: signals.OperationDelete,
	})
}
