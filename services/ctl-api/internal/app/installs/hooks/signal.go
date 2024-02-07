package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/signals"
	"go.uber.org/zap"
)

func (a *Hooks) sendSignal(ctx context.Context, installID string, signal signals.Signal) {
	err := a.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		signals.EventLoopWorkflowID(installID),
		"",
		installID,
		signal,
	)
	a.l.Debug("event workflow signaled", zap.Error(err))
}
