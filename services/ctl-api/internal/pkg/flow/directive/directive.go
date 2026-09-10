// Package directive defines typed directive constants for the workflow step and
// group execution system. Directives control the flow of execution: whether to
// continue, stop, retry, or skip.
//
// There are two levels:
//   - Step directives: written by step Execute() into the step's ResultDirective
//   - Group directives: written by group Execute() into the group's ResultDirective
//
// The flow executor reads group directives to decide workflow-level behavior.
// The group's sequential loop reads step directives to decide group-level behavior.
package directive

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

// Step is the typed directive written by step Execute() into the step's ResultDirective.
// The group reads this after the step's queue signal completes.
//
// Retry semantics: a failing step first consumes its auto-retry budget —
// StepAutoRetry / StepAutoRetryGroup clone and re-run with no human involved.
// Once auto-retries are exhausted (or the error hints auto-retry won't help),
// the step parks on StepAwaitManualRetry: Execute() blocks and the workflow
// sits at StatusFailedPendingRetry until a user retries or skips. The manual
// retry handler converges back onto the same StepAutoRetry / StepAutoRetryGroup
// directives the auto path uses — the only difference is who triggered it.
//
// StepAutoRetry, StepAutoRetryGroup, StepStop, StepSkipGroup are terminal:
// the group acts on them. StepAwaitManualRetry and StepAwaitApproval are
// non-terminal: Execute() is still blocking and the group never sees them.
type Step string

const (
	// StepContinue means the step succeeded. The group proceeds to the next step.
	StepContinue Step = "continue"

	// StepStop means the step failed terminally. The group stops and the workflow errors.
	StepStop Step = "stop"

	// StepAutoRetry means the step should be cloned and retried now, without
	// human involvement. Written when auto-retries remain in the budget; the
	// manual-retry handler writes the same directive after a user retries.
	StepAutoRetry Step = "retry"

	// StepAutoRetryGroup means the entire group should be cloned and retried
	// now, without human involvement. The group propagates this to the flow,
	// which handles group-level cloning.
	StepAutoRetryGroup Step = "retry-group"

	// StepSkipGroup means the remaining steps in the group should be skipped.
	// Used when a plan detects no changes (noop).
	StepSkipGroup Step = "skip-group"

	// StepAwaitApproval means the step is awaiting user approval. Execute()
	// blocks internally until the approval is resolved.
	StepAwaitApproval Step = "await-approval"

	// StepAwaitManualRetry means auto-retries are exhausted but the total retry
	// budget (auto + manual) is not. Execute() parks — blocking in the Temporal
	// workflow with the flow at StatusFailedPendingRetry — until the user
	// retries or skips, or the park ceiling abandons the step.
	StepAwaitManualRetry Step = "await-retry"
)

// IsTerminal returns true if the directive represents a completed step that the
// group should act on. Non-terminal directives (await-approval, await-retry)
// mean Execute() is still blocking — the group should not see these.
func (d Step) IsTerminal() bool {
	switch d {
	case StepContinue, StepStop, StepAutoRetry, StepAutoRetryGroup, StepSkipGroup:
		return true
	default:
		return false
	}
}

// StepResult carries the directive along with metadata that controls how the
// group and flow set statuses on remaining steps. This allows the step to
// communicate context (e.g., "denied" vs "error") to the group/flow.
type StepResult struct {
	// Directive is the typed step directive.
	Directive Step

	// Reason is a human-readable description of why this directive was issued.
	// Written to the step's status and propagated to remaining step metadata.
	// Examples: "approval denied", "max retries exhausted", "noop plan".
	Reason string

	// SiblingStatus is the status to apply to remaining steps in the SAME group
	// when the directive is StepStop or StepSkipGroup. Defaults to StatusDiscarded.
	SiblingStatus app.Status

	// FutureStatus is the status to apply to steps in FUTURE groups when the
	// directive is StepStop. If empty, defaults to StatusNotAttempted.
	FutureStatus app.Status
}

// NewStepResult creates a StepResult with sensible defaults.
func NewStepResult(d Step) StepResult {
	return StepResult{
		Directive:     d,
		SiblingStatus: app.StatusDiscarded,
		FutureStatus:  app.StatusNotAttempted,
	}
}

// Group is the typed directive written by group Execute() into the group's ResultDirective.
// The flow executor reads this after the group's queue signal completes.
type Group string

const (
	// GroupContinue means the group completed. The flow proceeds to the next group.
	GroupContinue Group = "continue"

	// GroupStop means the group stopped. The flow marks remaining groups as discarded.
	GroupStop Group = "stop"

	// GroupRetryGroup means the group should be cloned and retried.
	// The flow creates a new group with cloned steps and re-dispatches.
	GroupRetryGroup Group = "retry-group"

	// GroupSkipGroup means the group was skipped (e.g., noop plan).
	// The flow proceeds to the next group.
	GroupSkipGroup Group = "skip-group"

	// GroupAwaitApproval means the group is awaiting approval.
	GroupAwaitApproval Group = "await-approval"
)

// MetadataKey is the key used to store the directive in status metadata.
const MetadataKey = "directive"
