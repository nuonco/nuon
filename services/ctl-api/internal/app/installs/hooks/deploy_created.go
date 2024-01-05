package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/signals"
)

func (i *Hooks) InstallDeployCreated(ctx context.Context, installID, deployID string) {
	i.sendSignal(ctx, installID, signals.Signal{
		Operation: signals.OperationDeploy,
		DeployID:  deployID,
	})
}
