package workflowstepapprovalresponse

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// SignalType is the queue signal type for forwarding a workflow-step approval
// response to the running install workflow.
//
// This Nuon Signal wraps the Temporal "approve-step" update so the approval
// becomes a first-class operation: it's persisted in queue_signals, retried by
// the queue, surfaced in the dashboard, and emits webhook lifecycle events.
const SignalType signal.SignalType = "workflow-step-approval-response"

type Signal struct {
	InstallID          string                       `json:"install_id"`
	InstallWorkflowID  string                       `json:"install_workflow_id"`
	WorkflowStepID     string                       `json:"workflow_step_id"`
	ApprovalID         string                       `json:"approval_id"`
	ApprovalResponseID string                       `json:"approval_response_id"`
	ResponseType       app.WorkflowStepResponseType `json:"response_type"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) AutoRetry() bool { return true }
func (s *Signal) MaxRetries() int { return 5 }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	installID := &s.InstallID
	if s.InstallID == "" {
		installID = nil
	}
	// Expose workflow + step identity so lifecycle hooks (e.g. webhook) can
	// emit workflow_step.approval.v1 events without needing to dig into the
	// approval-specific signal payload. OwnerID/OwnerType point at the
	// install (the queue's owner), matching the convention used by
	// execute-workflow / execute-workflow-step lifecycle events.
	return signal.SignalLifecycleContext{
		InstallID:  installID,
		Operation:  "workflow-step-approval-response",
		WorkflowID: s.InstallWorkflowID,
		StepID:     s.WorkflowStepID,
		OwnerID:    s.InstallID,
		OwnerType:  "installs",
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return errors.New("install_id is required")
	}
	if s.InstallWorkflowID == "" {
		return errors.New("install_workflow_id is required")
	}
	if s.WorkflowStepID == "" {
		return errors.New("workflow_step_id is required")
	}
	if s.ApprovalResponseID == "" {
		return errors.New("approval_response_id is required")
	}
	if s.ResponseType == "" {
		return errors.New("response_type is required")
	}

	if _, err := activities.AwaitGetByInstallID(ctx, s.InstallID); err != nil {
		return errors.Wrap(err, "install not found")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if _, err := activities.AwaitApprovePlan(ctx, &activities.ApprovePlanRequest{
		InstallWorkflowID:  s.InstallWorkflowID,
		StepID:             s.WorkflowStepID,
		ApprovalResponseID: s.ApprovalResponseID,
		ResponseType:       s.ResponseType,
	}); err != nil {
		return errors.Wrap(err, "unable to forward approval to install workflow")
	}
	return nil
}
