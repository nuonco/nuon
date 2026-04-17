package example

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const FakeStepSignalType signal.SignalType = "fake-step-signal"

// FakeStepSignal is a composable test signal that can opt into any combination
// of the signal interfaces defined in signal/interfaces.go.
//
// Each boolean flag opts the signal into the corresponding interface behavior.
// The execute-workflow-step signal checks these interfaces via type assertion
// and drives behavior accordingly.
type FakeStepSignal struct {
	// Core behavior: "pass", "fail", "slow"
	Behavior string `json:"behavior"`

	// Interface opt-ins
	EnableAutoRetry        bool `json:"enable_auto_retry,omitempty"`
	CustomMaxRetries       int  `json:"custom_max_retries,omitempty"`
	EnableNoOpCheck        bool `json:"enable_noop_check,omitempty"`
	EnablePolicyEvaluation bool `json:"enable_policy_evaluation,omitempty"`

	// CloneStepNames, if non-empty, opts into SignalWithCloneSteps.
	CloneStepNames []string `json:"clone_step_names,omitempty"`

	// Step context (set by ApplyStepContext)
	WorkflowStepID string `json:"workflow_step_id,omitempty"`
	FlowID         string `json:"flow_id,omitempty"`

	// Observable state
	CancelCallbackHit bool `json:"cancel_callback_hit,omitempty"`
}

// Verify interface compliance at compile time.
var _ signal.Signal = (*FakeStepSignal)(nil)
var _ signal.SignalWithAutoRetry = (*FakeStepSignal)(nil)
var _ signal.SignalWithMaxRetries = (*FakeStepSignal)(nil)
var _ signal.SignalWithNoOpCheck = (*FakeStepSignal)(nil)
var _ signal.SignalWithPolicyEvaluation = (*FakeStepSignal)(nil)
var _ signal.SignalWithCancel = (*FakeStepSignal)(nil)
var _ signal.SignalWithCloneSteps = (*FakeStepSignal)(nil)

func (s *FakeStepSignal) Type() signal.SignalType {
	return FakeStepSignalType
}

func (s *FakeStepSignal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
	s.FlowID = flowID
}

func (s *FakeStepSignal) Validate(_ workflow.Context) error {
	return nil
}

func (s *FakeStepSignal) Execute(ctx workflow.Context) error {
	switch s.Behavior {
	case "pass":
		return nil
	case "fail":
		return errors.New("intentional test failure")
	case "slow":
		return workflow.Await(ctx, func() bool {
			return ctx.Err() != nil
		})
	default:
		return nil
	}
}

// SignalWithAutoRetry
func (s *FakeStepSignal) AutoRetry() bool {
	return s.EnableAutoRetry
}

// SignalWithMaxRetries
func (s *FakeStepSignal) MaxRetries() int {
	if s.CustomMaxRetries > 0 {
		return s.CustomMaxRetries
	}
	return signal.DefaultMaxRetries
}

// SignalWithNoOpCheck
func (s *FakeStepSignal) IsNoOpCheckable() bool {
	return s.EnableNoOpCheck
}

// SignalWithPolicyEvaluation
func (s *FakeStepSignal) RequiresPolicyEvaluation() bool {
	return s.EnablePolicyEvaluation
}

// SignalWithCancel
func (s *FakeStepSignal) Cancel(_ workflow.Context) error {
	s.CancelCallbackHit = true
	return nil
}

// SignalWithCloneSteps — returns custom clone step definitions when CloneStepNames is set.
// When empty, the execute-workflow-step falls through to default single-step clone.
func (s *FakeStepSignal) CloneSteps(originalStepName string) []signal.CloneStepDef {
	if len(s.CloneStepNames) == 0 {
		return nil
	}
	defs := make([]signal.CloneStepDef, len(s.CloneStepNames))
	for i, name := range s.CloneStepNames {
		defs[i] = signal.CloneStepDef{
			Signal: &FakeStepSignal{Behavior: "pass"},
			Name:   name,
		}
	}
	return defs
}
