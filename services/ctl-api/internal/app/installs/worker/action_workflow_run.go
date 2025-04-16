package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 30s
func (w *Workflows) ActionWorkflowRun(ctx workflow.Context, sreq signals.RequestSignal) error {
	return w.executeActionWorkflowRun(ctx, sreq.ID, sreq.ActionWorkflowRunID)
}

