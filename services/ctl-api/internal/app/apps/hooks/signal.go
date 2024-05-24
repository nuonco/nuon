package hooks

import (
	"context"

	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker/signals"
)

func (a *Hooks) sendSignal(ctx context.Context, appID string, signal signals.Signal) {
	err := a.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		worker.EventLoopWorkflowID(appID),
		"",
		appID,
		signal,
	)
	a.l.Debug("event workflow signaled", zap.Error(err))
}
