package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/signals"
)

func (a *Hooks) Deleted(ctx context.Context, appID string) {
	a.sendSignal(ctx, appID, signals.Signal{
		Operation: signals.OperationDeprovision,
	})
}
