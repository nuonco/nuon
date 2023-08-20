package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker"
)

func (h *Hooks) BuildCreated(ctx context.Context, componentID, buildID string) {
	h.sendSignal(ctx, componentID, worker.Signal{
		DryRun:    h.cfg.DevEnableWorkersDryRun,
		Operation: worker.OperationBuild,
		BuildID:   buildID,
	})
}
