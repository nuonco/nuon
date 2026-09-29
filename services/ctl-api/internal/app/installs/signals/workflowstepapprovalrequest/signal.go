package workflowstepapprovalrequest

import (
	goerrors "errors"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "workflow-step-approval-request"

const installWorkflowStepsOwnerType = "install_workflow_steps"

type Signal struct {
	InstallID         string `json:"install_id"`
	InstallWorkflowID string `json:"install_workflow_id"`
	WorkflowStepID    string `json:"workflow_step_id"`

	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`

	RunnerJobID string `json:"runner_job_id,omitempty"`

	ApprovalType app.WorkflowStepApprovalType `json:"approval_type"`

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

func Dispatch(ctx workflow.Context, sig *Signal) error {
	cb := callback.New(ctx, sig.WorkflowStepID)
	_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         sig.InstallID,
		OwnerType:       "installs",
		QueueName:       queuenames.InstallApprovalsQueueName,
		Signal:          sig,
		SignalOwnerID:   sig.WorkflowStepID,
		SignalOwnerType: installWorkflowStepsOwnerType,
		Callback:        cb,
	})
	if err != nil {
		return errors.Wrap(err, "unable to enqueue workflow-step-approval-request signal")
	}

	if _, err := callback.AwaitWithTimeout(ctx, cb, callback.HumanGatedTimeout); err != nil {
		if goerrors.Is(err, callback.ErrAwaitTimeout) {
			return errors.Wrap(err, "no approval response received")
		}
		return errors.Wrap(err, "workflow-step-approval-request signal failed")
	}
	return nil
}
