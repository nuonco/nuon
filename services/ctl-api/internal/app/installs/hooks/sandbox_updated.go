package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (i *Hooks) SandboxUpdated(ctx context.Context, installID string) {
	i.l.Info("sandbox was updated", zap.String("id", installID))
}
