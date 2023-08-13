package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (i *Hooks) Created(ctx context.Context, installID string) {
	i.l.Info("install created", zap.String("id", installID))
}
