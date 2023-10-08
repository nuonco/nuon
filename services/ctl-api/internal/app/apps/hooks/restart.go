package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (a *Hooks) Restart(ctx context.Context, appID string, sandboxMode bool) {
	if err := a.startEventLoop(ctx, appID, sandboxMode); err != nil {
		a.l.Error("error starting event loop",
			zap.String("app-id", appID),
			zap.Error(err),
		)
		return
	}
}
