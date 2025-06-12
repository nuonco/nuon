package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 2m
func (w *Workflows) WorkflowApproveAll(ctx workflow.Context, sreq signals.RequestSignal) error {
	// require approvals is an approval step so we need to create an approval for this step
	_, err := activities.AwaitCreateStepApproval(ctx, &activities.CreateStepApprovalRequest{
		OwnerID:   sreq.FlowID,
		OwnerType: "install_workflows",
		StepID:    sreq.FlowStepID,
		Type:      app.ApproveAllApprovalType,
	})
	if err != nil {
		return err
	}

	return nil
}
