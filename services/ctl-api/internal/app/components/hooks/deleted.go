package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker"
)

func (a *Hooks) Deleted(ctx context.Context, componentID string) {
	a.sendSignal(ctx, componentID, worker.Signal{
		DryRun:    a.cfg.DevEnableWorkersDryRun,
		Operation: worker.OperationDelete,
	})
}
