package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/signals"
)

func (i *Hooks) DeployComponents(ctx context.Context, installID string) {
	i.sendSignal(ctx, installID, signals.Signal{
		Operation: signals.OperationDeployComponents,
		Async:     true,
	})
}
