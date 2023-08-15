package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (h *Hooks) Created(ctx context.Context, componentID string) {
	h.l.Info("component created", zap.String("component-id", componentID))
}
