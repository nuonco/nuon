package executeworkflowstep

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "execute-workflow-step"

// Policy evaluation metadata keys
const (
	DenyViolationsKey  = "deny_violations"
	WarnViolationsKey  = "warn_violations"
	PassedPolicyIDsKey = "passed_policy_ids"
)

// stepSignal extracts the underlying signal.Signal from a workflow step's QueueSignal,
// returning nil if either is absent. Used for interface type assertions on optional
// signal capabilities (e.g. SignalWithNoOpCheck, SignalWithPolicyEvaluation).
func stepSignal(step *app.WorkflowStep) signal.Signal {
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		return step.QueueSignal.Signal
	}
	return nil
}

// Signal encapsulates the full lifecycle of executing a single workflow step.
// When dispatched to the step queue (e.g. install-workflow-steps), it fetches all
// necessary state from the database and runs the complete step execution flow:
// status updates, inner signal dispatch, approval handling, policy checks, and
// noop plan detection.
type Signal struct {
	StepID     string `json:"step_id"`
	WorkflowID string `json:"workflow_id"`

	// OwnerID is the entity that owns the queues (e.g. install ID).
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`

	// TargetQueueName is the queue where the inner signal (the actual step signal)
	// gets enqueued for execution (e.g. "install-signals").
	TargetQueueName string `json:"target_queue_name"`

	// innerQueueSignalID tracks the currently executing inner signal so that
	// Cancel() can propagate cancellation to it. Set during executeInnerSignal,
	// read during Cancel. Safe because Temporal workflows are single-threaded.
	innerQueueSignalID string

	// finished is set when Execute() completes (success or error). The
	// step-finished update handler blocks until this is true, then returns the
	// step's final status and directive.
	finished bool

	// retried is set by the "was-retried" update handler when the step-group
	// decides to retry this step. It unblocks waitForApprovalResponse so the
	// step can exit cleanly.
	retried bool

	// approved is set by the "approve-plan" update handler to reactively unblock
	// approval waiting instead of polling.
	approved bool

	canceled bool

	// approvalResponseID and approvalResponseType are set by the "approve-plan"
	// update handler to pass the response through without re-fetching from DB.
	approvalResponseID   string
	approvalResponseType string
}

var (
	_ signal.Signal                   = (*Signal)(nil)
	_ signal.SignalWithCancel         = (*Signal)(nil)
	_ signal.SignalWithUpdateHandlers = (*Signal)(nil)
)

// RegisterUpdateHandlers registers step-level update handlers on the handler workflow.
// Called by the handler framework after initializeState(), before Execute().
func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "is-retryable",
		s.isRetryableHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "create-step-retry",
		s.createStepRetryHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "approve-plan",
		s.approvePlanHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "step-finished",
		s.stepFinishedHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "was-retried",
		s.wasRetriedHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "cancel-step",
		s.cancelStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	return nil
}

func (s *Signal) SleepAfter() time.Duration { return 500 * time.Millisecond }

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.StepID == "" {
		return errors.New("step_id is required")
	}
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}
	if s.OwnerID == "" {
		return errors.New("owner_id is required")
	}
	if s.OwnerType == "" {
		return errors.New("owner_type is required")
	}
	if s.TargetQueueName == "" {
		return errors.New("target_queue_name is required")
	}
	return nil
}
