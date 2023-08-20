package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker"
	"go.uber.org/zap"
)

const (
	defaultNamespace string = "components"
)

func (a *Hooks) sendSignal(ctx context.Context, componentID string, signal worker.Signal) {
	err := a.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		worker.EventLoopWorkflowID(componentID),
		"",
		componentID,
		signal,
	)
	a.l.Info("event workflow signaled", zap.Error(err))
}
