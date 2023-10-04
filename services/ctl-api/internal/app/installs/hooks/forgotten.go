package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker"
)

func (i *Hooks) Forgotten(ctx context.Context, installID string) {
	i.sendSignal(ctx, installID, worker.Signal{
		DryRun:    i.cfg.DevEnableWorkersDryRun,
		Operation: worker.OperationForgotten,
	})
}
