package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/worker"
	"go.uber.org/zap"
)

func (a *Hooks) sendSignal(ctx context.Context, appID string, signal worker.Signal) {
	err := a.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		worker.EventLoopWorkflowID(appID),
		"",
		appID,
		signal,
	)
	a.l.Debug("event workflow signaled", zap.Error(err))
}
