package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/signals"
	"go.uber.org/zap"
)

const (
	defaultNamespace string = "components"
)

func (a *Hooks) sendSignal(ctx context.Context, componentID string, signal signals.Signal) {
	err := a.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		signals.EventLoopWorkflowID(componentID),
		"",
		componentID,
		signal,
	)
	a.l.Debug("event workflow signaled", zap.Error(err))
}
