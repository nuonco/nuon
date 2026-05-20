package enqueuer

import (
	"go.temporal.io/sdk/workflow"
)

type SweepWorkflowRequest struct{}

type Workflows struct{}

func NewWorkflows() *Workflows {
	return &Workflows{}
}

func (w *Workflows) All() []any {
	return []any{
		w.EnqueuerSweep,
	}
}

// @temporal-gen-v2 workflow
// @task-queue "queue"
// @id-template enqueuer-sweep
func (w *Workflows) EnqueuerSweep(ctx workflow.Context, req SweepWorkflowRequest) (*SweepResponse, error) {
	return AwaitSweep(ctx, &SweepRequest{})
}
