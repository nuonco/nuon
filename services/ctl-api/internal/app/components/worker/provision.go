package worker

import (
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) provision(ctx workflow.Context, componentID string) error {
	w.updateStatus(ctx, componentID, StatusActive, "component is active")
	return nil
}
