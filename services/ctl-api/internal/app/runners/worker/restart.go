package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"go.temporal.io/sdk/workflow"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 1m
func (w *Workflows) Restart(ctx workflow.Context, sreq signals.RequestSignal) error {
	return nil
}
