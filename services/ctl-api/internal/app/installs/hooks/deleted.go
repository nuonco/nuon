package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (i *Hooks) Deleted(ctx context.Context, installID string) {
	i.l.Info("install deleted", zap.String("id", installID))
}
