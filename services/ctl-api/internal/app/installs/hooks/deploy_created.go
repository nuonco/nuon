package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (i *Hooks) InstallDeployCreated(ctx context.Context, installID, deployID string) {
	i.l.Info("install deploy created", zap.String("install-id", installID), zap.String("deploy-id", deployID))
}
