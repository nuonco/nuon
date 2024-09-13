package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"go.temporal.io/sdk/workflow"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) ForceDeprovision(ctx workflow.Context, sreq signals.RequestSignal) error {
	return w.deprovisionOrg(ctx, sreq.ID, sreq.SandboxMode)
}
