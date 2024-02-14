package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.uber.org/zap"
)

func (a *Hooks) Restart(ctx context.Context, releaseID string, orgType app.OrgType) {
	if err := a.startEventLoop(ctx, releaseID, orgType); err != nil {
		a.l.Error("error starting event loop",
			zap.String("release-id", releaseID),
			zap.Error(err),
		)
		return
	}
}
