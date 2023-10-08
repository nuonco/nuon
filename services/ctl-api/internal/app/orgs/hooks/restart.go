package hooks

import (
	"context"

	"go.uber.org/zap"
)

func (o *Hooks) Restart(ctx context.Context, orgID string, sandboxMode bool) {
	if err := o.startEventLoop(ctx, orgID, sandboxMode); err != nil {
		o.l.Error("error starting event loop",
			zap.String("org-id", orgID),
			zap.Error(err),
		)
		return
	}
}
