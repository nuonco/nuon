package enqueuer

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	sweepInterval     = 1 * time.Minute
	sweepMaxAliveTime = 24 * time.Hour
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
func (w *Workflows) EnqueuerSweep(ctx workflow.Context, req SweepWorkflowRequest) error {
	deadline := workflow.Now(ctx).Add(sweepMaxAliveTime)

	for workflow.Now(ctx).Before(deadline) {
		if err := workflow.Sleep(ctx, sweepInterval); err != nil {
			return err
		}

		if err := AwaitSweep(ctx, &SweepRequest{}); err != nil {
			workflow.GetLogger(ctx).Warn("sweep activity failed", "error", err)
		}
	}

	return workflow.NewContinueAsNewError(ctx, w.EnqueuerSweep, req)
}
