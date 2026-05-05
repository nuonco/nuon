package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/workflowstepapprovalrequest"
)

// @temporal-gen-v2 workflow
// @execution-timeout 1m
// @task-timeout 2m
func (w *Workflows) WorkflowApproveAll(ctx workflow.Context, sreq signals.RequestSignal) error {
	// require approvals is an approval step so we need to create an approval
	// for this step. Going through the workflow-step-approval-request signal
	// (instead of calling the CreateStepApproval activity directly) keeps the
	// row creation on the same queue/lifecycle/webhook plumbing as the
	// approval response, so consumers see a uniform shape for both sides of
	// the approval handshake.
	return workflowstepapprovalrequest.Dispatch(ctx, &workflowstepapprovalrequest.Signal{
		InstallID:         sreq.ID,
		InstallWorkflowID: sreq.FlowID,
		WorkflowStepID:    sreq.FlowStepID,
		OwnerID:           sreq.FlowID,
		OwnerType:         "install_workflows",
		ApprovalType:      app.ApproveAllApprovalType,
	})
}
