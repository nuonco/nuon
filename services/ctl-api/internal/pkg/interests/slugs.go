package interests

import "fmt"

// Slug prefix constants. Webhook payloads embed a top-level `interests` array
// of these slugs so consumers can route by prefix without re-implementing the
// classifier.
const (
	SlugPrefixResource = "resource:"
	SlugPrefixOp       = "op:"
	SlugPrefixOutcome  = "outcome:"
	SlugPrefixEvent    = "event:"
)

// Outcome slugs. outcome:completion is emitted on terminal events of any
// status; outcome:failures is emitted only when the terminal status is
// failed/cancelled.
const (
	SlugOutcomeCompletion = SlugPrefixOutcome + "completion"
	SlugOutcomeFailures   = SlugPrefixOutcome + "failures"
)

// Lifecycle event slugs (workflow / step transitions).
const (
	SlugEventLifecycleStarted   = SlugPrefixEvent + "lifecycle.started"
	SlugEventLifecycleSucceeded = SlugPrefixEvent + "lifecycle.succeeded"
	SlugEventLifecycleFailed    = SlugPrefixEvent + "lifecycle.failed"
	SlugEventLifecycleCancelled = SlugPrefixEvent + "lifecycle.cancelled"
)

// Approval handshake slugs. event:approval.response is emitted as a generic
// fallback when the specific approved/rejected outcome cannot be resolved
// (e.g. classify() called without a DB). The webhook hook upgrades the
// payload's interests array to the more specific form when it has the data.
const (
	SlugEventApprovalRequest          = SlugPrefixEvent + "approval.request"
	SlugEventApprovalResponse         = SlugPrefixEvent + "approval.response"
	SlugEventApprovalResponseApproved = SlugPrefixEvent + "approval.response.approved"
	SlugEventApprovalResponseRejected = SlugPrefixEvent + "approval.response.rejected"
)

// Drift detection slug. event:drift.detected is emitted by the drift-detected
// signal that fires from inside the plan-only check of a drift_run /
// drift_run_reprovision_sandbox workflow when the plan has changes (i.e.
// drift was actually found). Subscribers opt in via the per-resource
// `drift_detected` flag, which is gated independently of `outcome`.
const (
	SlugEventDriftDetected = SlugPrefixEvent + "drift.detected"
)

const (
	SlugEventRoleChange    = SlugPrefixEvent + "role.change"
	SlugEventInputsUpdated = SlugPrefixEvent + "inputs.updated"
)

// ResourceSlug returns "resource:<kind>".
func ResourceSlug(kind ResourceKind) string {
	return SlugPrefixResource + string(kind)
}

// OpSlug returns "op:<kind>.<op>".
func OpSlug(kind ResourceKind, op string) string {
	return fmt.Sprintf("%s%s.%s", SlugPrefixOp, kind, op)
}
