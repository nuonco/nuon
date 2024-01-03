package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/signals"
)

func (a *Hooks) Deleted(ctx context.Context, componentID string) {
	a.sendSignal(ctx, componentID, signals.Signal{
		Operation: signals.OperationDelete,
	})
}
