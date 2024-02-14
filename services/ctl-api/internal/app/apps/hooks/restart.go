package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.uber.org/zap"
)

func (a *Hooks) Restart(ctx context.Context, appID string, orgType app.OrgType) {
	if err := a.startEventLoop(ctx, appID, orgType); err != nil {
		a.l.Error("error starting event loop",
			zap.String("app-id", appID),
			zap.Error(err),
		)
		return
	}
}
