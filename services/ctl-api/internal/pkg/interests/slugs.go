package interests

import "fmt"

const (
	SlugPrefixResource = "resource:"
	SlugPrefixOp       = "op:"
	SlugPrefixOutcome  = "outcome:"
	SlugPrefixEvent    = "event:"
)

const (
	SlugOutcomeCompletion = SlugPrefixOutcome + "completion"
	SlugOutcomeFailures   = SlugPrefixOutcome + "failures"
)

const (
	SlugEventLifecycleStarted   = SlugPrefixEvent + "lifecycle.started"
	SlugEventLifecycleSucceeded = SlugPrefixEvent + "lifecycle.succeeded"
	SlugEventLifecycleFailed    = SlugPrefixEvent + "lifecycle.failed"
	SlugEventLifecycleCancelled = SlugPrefixEvent + "lifecycle.cancelled"
)

const (
	SlugEventApprovalRequest          = SlugPrefixEvent + "approval.request"
	SlugEventApprovalResponse         = SlugPrefixEvent + "approval.response"
	SlugEventApprovalResponseApproved = SlugPrefixEvent + "approval.response.approved"
	SlugEventApprovalResponseRejected = SlugPrefixEvent + "approval.response.rejected"
)

const (
	SlugEventDriftDetected = SlugPrefixEvent + "drift.detected"
)

const (
	SlugEventRoleChange      = SlugPrefixEvent + "role.change"
	SlugEventInputsUpdated   = SlugPrefixEvent + "inputs.updated"
	SlugEventConfigSynced    = SlugPrefixEvent + "config.synced"
	SlugEventRunnerUnhealthy = SlugPrefixEvent + "runner.unhealthy"
)

const (
	SlugEventComponentUnhealthy = SlugPrefixEvent + "component.unhealthy"
	SlugEventComponentRecovered = SlugPrefixEvent + "component.recovered"
	SlugEventInstallDegraded    = SlugPrefixEvent + "install.degraded"
)

func ResourceSlug(kind ResourceKind) string {
	return SlugPrefixResource + string(kind)
}

func OpSlug(kind ResourceKind, op string) string {
	return fmt.Sprintf("%s%s.%s", SlugPrefixOp, kind, op)
}
