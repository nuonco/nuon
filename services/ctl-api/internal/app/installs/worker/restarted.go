package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 1m
func (w *Workflows) Restarted(ctx workflow.Context, sreq signals.RequestSignal) error {
	return nil
}
