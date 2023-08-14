package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (a *Hooks) Created(ctx context.Context, id string) {
	a.l.Info("new app created", zap.String("id", id))
}
