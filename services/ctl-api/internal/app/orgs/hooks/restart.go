package hooks

import (
	"context"

	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/signals"
)

func (o *Hooks) Restart(ctx context.Context, orgID string, orgType app.OrgType) {
	if err := o.startEventLoop(ctx, orgID, orgType); err != nil {
		o.l.Error("error starting event loop",
			zap.String("org-id", orgID),
			zap.Error(err),
		)
		return
	}

	o.sendSignal(ctx, orgID, signals.Signal{
		Operation: signals.OperationRestart,
	})
}
