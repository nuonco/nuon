package hooks

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker"
	"go.uber.org/zap"
)

func (h *Hooks) sendSignal(ctx context.Context, orgID string, signal worker.Signal) {
	err := h.client.SignalWorkflowInNamespace(ctx,
		defaultNamespace,
		worker.EventLoopWorkflowID(orgID),
		"",
		orgID,
		signal,
	)
	h.l.Debug("event workflow signaled", zap.Error(err))
}
