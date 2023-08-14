package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (a *Hooks) Deleted(ctx context.Context, id string) {
	a.l.Info("app deleted", zap.String("id", id))
}
