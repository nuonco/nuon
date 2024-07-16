package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (w *Workflows) provision(ctx workflow.Context, componentID string) error {
	return nil
}

func (w *Workflows) created(ctx workflow.Context, componentID string) error {
	w.updateStatus(ctx, componentID, app.ComponentStatusActive, "component is active")
	return nil
}
