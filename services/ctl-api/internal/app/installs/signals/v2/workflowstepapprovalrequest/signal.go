package workflowstepapprovalrequest

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// SignalType is the queue signal type for creating a WorkflowStepApproval row.
//
// This Nuon Signal wraps the row-creation that previously happened directly
// via the CreateStepApproval activity inside running workflows. Wrapping it
// turns approval creation into a first-class operation: it's persisted in
// queue_signals, visible in the dashboard, retried by the queue, and emits
// webhook lifecycle events. Together with the workflow-step-approval-response
// signal it gives both sides of the approval handshake a uniform shape.
const SignalType signal.SignalType = "workflow-step-approval-request"

// installSignalsQueueName mirrors the constant in
// services/ctl-api/internal/app/installs/helpers. Duplicated here as a
// literal to avoid an import cycle (helpers imports signals via fx wiring).
const installSignalsQueueName = "install-signals"

// installWorkflowStepsOwnerType matches the polymorphic type used by
// QueueSignal records that originate from a workflow step.
const installWorkflowStepsOwnerType = "install_workflow_steps"

type Signal struct {
	InstallID         string `json:"install_id"`
	InstallWorkflowID string `json:"install_workflow_id"`
	WorkflowStepID    string `json:"workflow_step_id"`

	// OwnerID / OwnerType are the polymorphic owner stored on the resulting
	// WorkflowStepApproval row. They reference the entity the approval is
	// "for" (e.g. an install_deploys, install_sandbox_runs, or
	// install_workflows row).
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`

	// RunnerJobID, when set, links the approval back to the runner job that
	// produced its plan. May be empty for approvals that don't have an
	// underlying runner job (e.g. workflow-level approve-all).
	RunnerJobID string `json:"runner_job_id,omitempty"`

	ApprovalType app.WorkflowStepApprovalType `json:"approval_type"`

	// Plan is the rendered plan contents stored on the approval row. When
	// empty, CreateStepApproval falls back to fetching it from the runner
	// job results.
	Plan string `json:"plan,omitempty"`
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
		Operation:  "workflow-step-approval-request",
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
	if s.WorkflowStepID == "" {
		return errors.New("workflow_step_id is required")
	}
	if s.OwnerID == "" {
		return errors.New("owner_id is required")
	}
	if s.OwnerType == "" {
		return errors.New("owner_type is required")
	}
	if s.ApprovalType == "" {
		return errors.New("approval_type is required")
	}

	if _, err := activities.AwaitGetByInstallID(ctx, s.InstallID); err != nil {
		return errors.Wrap(err, "install not found")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if _, err := activities.AwaitCreateStepApproval(ctx, &activities.CreateStepApprovalRequest{
		OwnerID:     s.OwnerID,
		OwnerType:   s.OwnerType,
		RunnerJobID: s.RunnerJobID,
		StepID:      s.WorkflowStepID,
		Type:        s.ApprovalType,
		Plan:        s.Plan,
	}); err != nil {
		return errors.Wrap(err, "unable to create workflow step approval")
	}
	return nil
}

// Dispatch enqueues a workflow-step-approval-request signal onto the
// install-signals queue and waits for it to reach a terminal phase. Calling
// this from inside a running workflow replaces the direct
// activities.AwaitCreateStepApproval call so approval row creation flows
// through the same queue/lifecycle/webhook plumbing as the approval response.
//
// The caller is responsible for populating InstallID, InstallWorkflowID,
// WorkflowStepID, OwnerID, OwnerType, and Type on sig before calling. The
// signal is enqueued onto the install-signals queue with the workflow step
// as its owner so it shows up in the same place in the dashboard as other
// step-scoped signals.
func Dispatch(ctx workflow.Context, sig *Signal) error {
	resp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         sig.InstallID,
		OwnerType:       "installs",
		QueueName:       installSignalsQueueName,
		Signal:          sig,
		SignalOwnerID:   sig.WorkflowStepID,
		SignalOwnerType: installWorkflowStepsOwnerType,
	})
	if err != nil {
		return errors.Wrap(err, "unable to enqueue workflow-step-approval-request signal")
	}

	if _, err := client.AwaitAwaitSignal(ctx, resp.QueueSignalID); err != nil {
		return errors.Wrap(err, "workflow-step-approval-request signal failed")
	}
	return nil
}
