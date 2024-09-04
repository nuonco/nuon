package worker

import "go.temporal.io/sdk/workflow"

func (w *Workflows) forceDeprovision(ctx workflow.Context, orgID string, sandboxMode bool) error {
	return w.deprovisionOrg(ctx, orgID, sandboxMode)
}
