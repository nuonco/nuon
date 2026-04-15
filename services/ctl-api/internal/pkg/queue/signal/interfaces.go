package signal

import "go.temporal.io/sdk/workflow"

// This file defines all optional extension interfaces that signals can implement
// to declare capabilities and attributes. The conductor and step workflow check
// for these interfaces at runtime to drive behavior.
//
// Pattern: each interface is checked via type assertion on the signal.Signal
// stored in step.QueueSignal.Signal. Implementing an interface opts a signal
// into the corresponding behavior without modifying conductor code.

// ---------------------------------------------------------------------------
// Step Retry & Clone
// ---------------------------------------------------------------------------

// CloneStepDef describes a step to create when cloning a signal's step for retry.
type CloneStepDef struct {
	Signal        Signal
	Name          string
	ExecutionType string // maps to app.WorkflowStepExecutionType (string to avoid importing app)
}

// SignalWithCloneSteps is implemented by signals whose steps require prerequisite
// steps when retried. For example, an apply signal may return a plan step + apply step
// so that retrying an apply re-runs the plan first.
type SignalWithCloneSteps interface {
	CloneSteps(originalStepName string) []CloneStepDef
}

// SignalWithMaxRetries is implemented by signals that declare a custom maximum retry count.
// If not implemented, the default max retry count (DefaultMaxRetries) is used.
type SignalWithMaxRetries interface {
	MaxRetries() int
}

// DefaultMaxRetries is the default maximum retry count for steps that don't implement SignalWithMaxRetries.
const DefaultMaxRetries = 10

// SignalWithAutoRetry is implemented by signals that should automatically retry
// on failure without requiring user intervention. The conductor checks this after
// a step error and triggers a retry if the retry budget hasn't been exhausted.
type SignalWithAutoRetry interface {
	AutoRetry() bool
}

// ---------------------------------------------------------------------------
// Step Execution Capabilities
// ---------------------------------------------------------------------------

// SignalWithNoOpCheck is implemented by approval-type signals whose plan can
// have zero changes (noop). When the conductor detects a noop plan, it auto-skips
// both the approval step and its paired apply step.
type SignalWithNoOpCheck interface {
	IsNoOpCheckable() bool
}

// SignalWithPolicyEvaluation is implemented by signals whose steps require
// policy evaluation before the approval proceeds. The conductor runs policy
// checks only when this interface is present and returns true.
type SignalWithPolicyEvaluation interface {
	RequiresPolicyEvaluation() bool
}

// ---------------------------------------------------------------------------
// Step Lifecycle
// ---------------------------------------------------------------------------

// SignalWithSkipCleanup is implemented by signals that need to perform cleanup
// when their step is skipped or denied with "skip dependents". For example,
// sandbox plan signals mark all component deploy steps as skipped.
// The conductor calls OnSkipped after marking the step itself.
type SignalWithSkipCleanup interface {
	OnSkipped(ctx workflow.Context) error
}

// ---------------------------------------------------------------------------
// Queue & Scheduling
// ---------------------------------------------------------------------------

// SignalWithQueue is implemented by signals that target a specific workflow queue
// instead of the conductor's default target queue. The conductor reads this to
// route the inner signal to the correct queue.
type SignalWithQueue interface {
	Queue() string
}

// SignalWithParallelizable is implemented by signals whose steps can execute
// in parallel with other steps in the same group. By default, steps within
// a group execute sequentially. When true, the conductor may dispatch multiple
// steps concurrently.
type SignalWithParallelizable interface {
	IsParallelizable() bool
}
