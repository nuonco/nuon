package worker

import "go.temporal.io/sdk/workflow"

func (w *Workflows) created(ctx workflow.Context, runnerID string) error {
	return nil
}
