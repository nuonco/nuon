package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/signals"
)

func (o *Hooks) Reprovision(ctx context.Context, appID string) {
	o.l.Info("sending signal")
	o.sendSignal(ctx, appID, signals.Signal{
		Operation: signals.OperationReprovision,
	})
}
