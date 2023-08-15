package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (h *Hooks) BuildCreated(ctx context.Context, componentID string) {
	h.l.Info("component build created", zap.String("component-id", componentID))
}
