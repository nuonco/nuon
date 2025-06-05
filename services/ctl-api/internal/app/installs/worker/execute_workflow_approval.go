package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) WorkflowApproval(ctx workflow.Context, sreq signals.RequestSignal) error {
	_, err := activities.AwaitCreateInstallWorkflowApproval(ctx, &activities.CreateInstallWorkflowApprovalRequest{
		InstallWorkflowStepID: sreq.WorkflowStepID,
	})
	if err != nil {
		return nil
		// return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
	}

	return nil
}
