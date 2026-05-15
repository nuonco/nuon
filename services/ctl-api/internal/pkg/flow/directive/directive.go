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
type Step string

const (
	// StepContinue means the step succeeded. The group proceeds to the next step.
	StepContinue Step = "continue"

	// StepStop means the step failed terminally. The group stops and the workflow errors.
	StepStop Step = "stop"

	// StepRetry means the step should be cloned and retried individually.
	// The group creates a clone and picks it up on the next loop iteration.
	StepRetry Step = "retry"

	// StepRetryGroup means the entire group should be cloned and retried.
	// The group propagates this to the flow, which handles group-level cloning.
	StepRetryGroup Step = "retry-group"

	// StepSkipGroup means the remaining steps in the group should be skipped.
	// Used when a plan detects no changes (noop).
	StepSkipGroup Step = "skip-group"

	// StepAwaitApproval means the step is awaiting user approval. Execute()
	// blocks internally until the approval is resolved.
	StepAwaitApproval Step = "await-approval"

	// StepAwaitRetry means auto-retries are exhausted but manual retries remain.
	// Execute() blocks internally until the user retries or skips.
	StepAwaitRetry Step = "await-retry"
)

// IsTerminal returns true if the directive represents a completed step that the
// group should act on. Non-terminal directives (await-approval, await-retry)
// mean Execute() is still blocking — the group should not see these.
func (d Step) IsTerminal() bool {
	switch d {
	case StepContinue, StepStop, StepRetry, StepRetryGroup, StepSkipGroup:
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
