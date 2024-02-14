package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.uber.org/zap"
)

func (a *Hooks) Restart(ctx context.Context, installID string, orgType app.OrgType) {
	if err := a.startEventLoop(ctx, installID, orgType); err != nil {
		a.l.Error("unable to start event loop",
			zap.String("install-id", installID),
			zap.Error(err),
		)
		return
	}
}
