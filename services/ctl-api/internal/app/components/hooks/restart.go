package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (a *Hooks) Restart(ctx context.Context, componentID string, sandboxMode bool) {
	if err := a.startEventLoop(ctx, componentID, sandboxMode); err != nil {
		a.l.Error("error starting event loop",
			zap.String("component-id", componentID),
			zap.Error(err),
		)
		return
	}
}
