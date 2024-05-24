package hooks

import (
	"context"

	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/signals"
)

func (h *Hooks) sendSignal(ctx context.Context, orgID string, signal signals.Signal) {
	err := h.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		worker.EventLoopWorkflowID(orgID),
		"",
		orgID,
		signal,
	)
	h.l.Debug("event workflow signaled", zap.Error(err))
}
