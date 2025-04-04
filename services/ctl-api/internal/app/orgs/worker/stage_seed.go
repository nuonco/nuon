package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) StageSeed(ctx workflow.Context, sreq signals.RequestSignal) error {
	// provision the org runner and make sure it's running
	return nil
}
