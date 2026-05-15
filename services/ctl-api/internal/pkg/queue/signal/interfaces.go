package signal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

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

// SignalWithRetryCount is implemented by signals that need to know their retry
// index at execution time. When a step is executed, the conductor calls
// SetRetryCount with the step's RetryIndex and GroupRetryIdx before Execute().
type SignalWithRetryCount interface {
	SetRetryCount(retryIndex, groupRetryIndex int)
}

// SignalWithClone is the unified clone interface for retry. When a step is
// cloned for retry (either individual retry or group retry), the clone
// machinery calls Clone() if implemented. The signal returns one or more
// CloneStepDef entries:
//
//   - Plan signals return a single entry with a clean copy (e.g., DeployID reset)
//   - Apply signals return multiple entries (plan + apply) so retrying an apply
//     re-runs the plan first
//
// Clone() is also responsible for cleaning up old targets (e.g., marking the
// old deploy as retried). Signals that don't implement this interface are
// copied verbatim as a single step.
//
// The stepName parameter is the original step name (with retry suffixes stripped),
// used by multi-step clones to derive human-readable names.
type SignalWithClone interface {
	Clone(ctx workflow.Context, stepName string) ([]CloneStepDef, error)
}

// SignalWithCloneSteps is the old name for SignalWithClone. Keeping as an alias
// so grep for the old name still finds the right type.
type SignalWithCloneSteps = SignalWithClone

// SignalWithMaxRetries is implemented by signals that declare a custom maximum retry count.
// If not implemented, the default max retry count (DefaultMaxRetries) is used.
// This is the global ceiling for retries — when exhausted, no more retries of any kind.
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

// SignalWithMaxAutoRetries is implemented by signals that want a separate limit
// for automatic retries (as opposed to user-initiated manual retries).
// When auto-retries are exhausted the step is marked as failed, but the user
// can still manually retry up to MaxRetries().
// If not implemented, all retries up to MaxRetries() are automatic.
//
// Accepts workflow.Context so implementations can fetch configuration
// dynamically (e.g., from the component config) rather than baking values
// into the signal at step creation time.
type SignalWithMaxAutoRetries interface {
	MaxAutoRetries(ctx workflow.Context) int
}

// SignalWithRetryGroup is implemented by signals whose steps should retry as an
// entire group. When a retry is triggered for a step whose signal returns
// RetryGroup() == true, ALL steps sharing the same GroupIdx are cloned and
// re-executed, not just the individual step.
//
// Mutually exclusive with parallel group execution — when a group has
// Parallel=true, RetryGroup is not supported and individual step retries
// are used instead.
type SignalWithRetryGroup interface {
	RetryGroup() bool
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

// SignalWithSkipNoops is implemented by plan signals that can control whether
// noop plans are auto-skipped. When SkipNoops returns false, noop plans proceed
// through the normal approval flow instead of being auto-skipped.
// If not implemented, noops are always auto-skipped (backward compatible).
type SignalWithSkipNoops interface {
	SkipNoops(ctx workflow.Context) bool
}

// SignalWithAutoApproveOnPoliciesPassing is implemented by plan signals that
// should be auto-approved when all policies pass (no deny violations). The
// policy check calls this after evaluation — if it returns true and there are
// no deny violations, the step is auto-approved without waiting for user input.
type SignalWithAutoApproveOnPoliciesPassing interface {
	AutoApproveOnPoliciesPassing(ctx workflow.Context) bool
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
// Approval Response Lifecycle Hooks
// ---------------------------------------------------------------------------

// SignalWithOnApprove is called when an approval step receives an "approve" response.
// Signals can implement this to perform side effects after approval is granted.
type SignalWithOnApprove interface {
	OnApprove(ctx workflow.Context) error
}

// SignalWithOnRetry is called when an approval step receives a "retry plan" response.
// Signals can implement this to perform cleanup or preparation before the retry clone is created.
type SignalWithOnRetry interface {
	OnRetry(ctx workflow.Context) error
}

// SignalWithOnSkip is called when an approval step receives a "skip" response
// (either skip-current or skip-current-and-dependents).
// Distinct from SignalWithSkipCleanup.OnSkipped which is called by the conductor
// when programmatically skipping a step.
type SignalWithOnSkip interface {
	OnSkip(ctx workflow.Context) error
}

// SignalWithOnDeny is called when an approval step receives a "deny" response.
// Signals can implement this to perform cleanup when approval is denied.
type SignalWithOnDeny interface {
	OnDeny(ctx workflow.Context) error
}

// SignalWithSkipGroup is implemented by signals that want a "skip" approval
// response to skip the entire remaining group (DirectiveSkipGroup). When not
// implemented or when SkipGroup() returns false, a skip response only skips
// the current step and continues to the next step in the group (DirectiveContinue).
type SignalWithSkipGroup interface {
	SkipGroup() bool
}

// ---------------------------------------------------------------------------
// Step Generation
// ---------------------------------------------------------------------------

// SignalWithFetchSteps marks a signal as a step generator.
// The signal must register a "FetchSteps" update handler that returns workflow steps.
// The conductor enqueues the signal, sends the "FetchSteps" update, and uses the
// returned steps instead of the hardcoded Generators map.
type SignalWithFetchSteps interface {
	Signal
	SignalWithUpdateHandlers
}

// ---------------------------------------------------------------------------
// Timeouts
// ---------------------------------------------------------------------------

// SignalWithTimeout is implemented by signals that need a custom execution timeout.
// The returned duration is used as the ScheduleToCloseTimeout for the execute
// activity in the queue dispatcher. If not implemented, the default (2 hours) is used.
type SignalWithTimeout interface {
	Timeout() time.Duration
}

// SignalWithMaxInFlightAge is implemented by signals that should not be deduped
// indefinitely. When EmitSignal finds an in-flight signal (queued or in-progress)
// for the same emitter older than this age, the stale signal is marked failed
// and a fresh one is emitted. Without this interface, in-flight signals dedup
// forever — fine for user-awaited signals (recoverable via AwaitSignal fallback),
// but a permanent block for fire-and-forget emitter signals like health checks.
type SignalWithMaxInFlightAge interface {
	MaxInFlightAge() time.Duration
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
