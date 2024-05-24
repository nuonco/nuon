package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/signals"
)

func (a *Hooks) ConfigCreated(ctx context.Context, appID, appConfigID string) {
	a.sendSignal(ctx, appID, signals.Signal{
		Operation:   signals.OperationConfigCreated,
		AppConfigID: appConfigID,
	})
}
