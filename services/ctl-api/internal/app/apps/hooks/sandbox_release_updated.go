package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (a *Hooks) SandboxReleaseUpdated(ctx context.Context, appID string) {
	a.l.Info("app sandbox release updated", zap.String("id", appID))
}
