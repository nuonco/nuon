package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (a *Hooks) Restart(ctx context.Context, releaseID string, sandboxMode bool) {
	if err := a.startEventLoop(ctx, releaseID, sandboxMode); err != nil {
		a.l.Error("error starting event loop",
			zap.String("release-id", releaseID),
			zap.Error(err),
		)
		return
	}
}
