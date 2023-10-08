package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker"
)

func (o *Hooks) Reprovision(ctx context.Context, installID string) {
	o.l.Info("sending signal")
	o.sendSignal(ctx, installID, worker.Signal{
		Operation: worker.OperationReprovision,
	})
}
