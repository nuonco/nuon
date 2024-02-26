package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker"
)

func (a *Hooks) ConfigCreated(ctx context.Context, appID, appConfigID string) {
	a.sendSignal(ctx, appID, worker.Signal{
		Operation:   worker.OperationConfigCreated,
		AppConfigID: appConfigID,
	})
}
