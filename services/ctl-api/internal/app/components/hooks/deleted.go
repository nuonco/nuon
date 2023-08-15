package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (h *Hooks) Deleted(ctx context.Context, componentID string) {
	h.l.Info("component deleted", zap.String("component-id", componentID))
}
