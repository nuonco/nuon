package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker"
	"go.uber.org/zap"
)

func (a *Hooks) sendSignal(ctx context.Context, installID string, signal worker.Signal) {
	err := a.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		worker.EventLoopWorkflowID(installID),
		"",
		installID,
		signal,
	)
	a.l.Info("event workflow signaled", zap.Error(err))
}
