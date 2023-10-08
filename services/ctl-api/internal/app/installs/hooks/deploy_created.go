package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker"
)

func (i *Hooks) InstallDeployCreated(ctx context.Context, installID, deployID string) {
	i.sendSignal(ctx, installID, worker.Signal{
		Operation: worker.OperationDeploy,
		DeployID:  deployID,
	})
}
