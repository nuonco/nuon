package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/releases/worker"
	"go.uber.org/zap"
)

func (a *Hooks) sendSignal(ctx context.Context, releaseID string, signal worker.Signal) {
	err := a.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		worker.EventLoopWorkflowID(releaseID),
		"",
		releaseID,
		signal,
	)
	a.l.Debug("event workflow signaled", zap.Error(err))
}
