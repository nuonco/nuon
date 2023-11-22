package worker

import (
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) provision(ctx workflow.Context, appID string) error {
	w.updateStatus(ctx, appID, StatusActive, "component is active")
	return nil
}
