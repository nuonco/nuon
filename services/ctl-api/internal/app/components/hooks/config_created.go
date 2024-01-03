package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/signals"
)

func (h *Hooks) ConfigCreated(ctx context.Context, componentID string) {
	h.sendSignal(ctx, componentID, signals.Signal{
		Operation: signals.OperationQueueBuild,
		BuildID:   componentID,
	})
}
