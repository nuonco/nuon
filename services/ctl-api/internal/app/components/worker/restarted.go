package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 1m
func (w *Workflows) Restarted(ctx workflow.Context, sreq signals.RequestSignal) error {
	w.updateStatus(ctx, sreq.ID, app.ComponentStatusActive, "component is active")
	return nil
}
