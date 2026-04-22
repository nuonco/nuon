package helpers

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/worker/runnerstatussignals"
)

// SignalOrgProvisionForRunner wakes any org provision or reprovision workflow
// that is waiting on this runner's status. Used by runner status writers so
// org workflows react immediately instead of waiting for the next 10-second poll tick.
func (h *Helpers) SignalOrgProvisionForRunner(ctx context.Context, l *zap.Logger, runnerID, reason string) {
	var runner app.Runner
	if err := h.db.WithContext(ctx).Select("org_id").First(&runner, "id = ?", runnerID).Error; err != nil {
		l.Debug("unable to find runner for org provision signal", zap.String("runner_id", runnerID), zap.Error(err))
		return
	}

	wakeup := runnerstatussignals.WakeUp{Reason: reason}
	for _, wfID := range []string{
		runnerstatussignals.ProvisionID(runner.OrgID),
		runnerstatussignals.ReprovisionID(runner.OrgID),
	} {
		if err := h.tClient.SignalWorkflowInNamespace(
			ctx,
			runnerstatussignals.Namespace,
			wfID,
			"",
			runnerstatussignals.SignalName,
			wakeup,
		); err != nil {
			l.Debug("org provision signal skipped",
				zap.String("workflow_id", wfID),
				zap.String("reason", reason),
				zap.Error(err),
			)
		}
	}
}
