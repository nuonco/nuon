package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.uber.org/zap"
)

func (a *Hooks) Restart(ctx context.Context, componentID string, orgType app.OrgType) {
	if err := a.startEventLoop(ctx, componentID, orgType); err != nil {
		a.l.Error("error starting event loop",
			zap.String("component-id", componentID),
			zap.Error(err),
		)
		return
	}
}
