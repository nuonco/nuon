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

	// WorkflowType identifies the kind of workflow that owns this step. Set at
	// dispatch time from the in-scope *app.Workflow so the lifecycle hook can
	// emit workflow_step.lifecycle events without a DB lookup.
	WorkflowType string `json:"workflow_type,omitempty"`

	// OwnerID is the entity that owns the queues (e.g. install ID).
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`

	// OrgID / OrgName / OwnerName are pass-through fields stamped by the
	// parent execute-workflow signal and forwarded by the step group at
	// dispatch time. Exposed via LifecycleContext so workflow_step lifecycle
	// webhook payloads carry human-readable names without a per-event DB
	// lookup.
	OrgID     string `json:"org_id,omitempty"`
	OrgName   string `json:"org_name,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`

	// TargetQueueName is the queue where the inner signal (the actual step signal)
	// gets enqueued for execution (e.g. "install-signals").
	TargetQueueName string `json:"target_queue_name"`

	// TargetQueueID, when set, is used directly for the inner signal dispatch,
	// bypassing the OwnerID/TargetQueueName lookup.
	TargetQueueID string `json:"target_queue_id,omitempty"`

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
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithCancel           = (*Signal)(nil)
	_ signal.SignalWithUpdateHandlers   = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
)

// LifecycleContext exposes the step + workflow identity to lifecycle hooks so
// they can emit workflow_step.lifecycle.* webhook events without leaking the
// inner signal's operation taxonomy.
func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		OrgID:        s.OrgID,
		OrgName:      s.OrgName,
		StepID:       s.StepID,
		WorkflowID:   s.WorkflowID,
		WorkflowType: s.WorkflowType,
		OwnerID:      s.OwnerID,
		OwnerType:    s.OwnerType,
		OwnerName:    s.OwnerName,
	}
}

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
	// TargetQueueName is optional — when empty, the inner signal is dispatched
	// to the owner's default queue (used with per-step signal queue routing).
	return nil
}
