package workflowapproveall

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "workflow-approve-all"

type Signal struct {
	signal.Hooks
	InstallID      string `json:"install_id"`
	WorkflowStepID string `json:"workflow_step_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}

	if s.WorkflowStepID == "" {
		return errors.New("workflow_step_id is required")
	}

	// Validate install exists
	_, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "install not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Create workflow approval (copied from worker/execute_workflow_approval.go)
	_, err := activities.AwaitCreateInstallWorkflowApproval(ctx, &activities.CreateInstallWorkflowApprovalRequest{
		InstallWorkflowStepID: s.WorkflowStepID,
	})
	if err != nil {
		return nil
		// Original code returns nil even on error, keeping that behavior
		// return w.handleStepErr(ctx, sreq.WorkflowStepID, err)
	}

	return nil
}
