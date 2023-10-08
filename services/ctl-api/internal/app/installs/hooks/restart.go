package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (a *Hooks) Restart(ctx context.Context, installID string, sandboxMode bool) {
	if err := a.startEventLoop(ctx, installID, sandboxMode); err != nil {
		a.l.Error("unable to start event loop",
			zap.String("install-id", installID),
			zap.Error(err),
		)
		return
	}
}
