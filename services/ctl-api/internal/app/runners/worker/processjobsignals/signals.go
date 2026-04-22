// Package processjobsignals defines the Temporal signal contract used to
// wake the ProcessJob workflow in response to external state changes.
//
// Before this package, ProcessJob polled runner-job / execution / runner
// state on a 1-second timer (see monitor_job_execution.go,
// start_job_execution.go). The polling generated ~15 history events per
// second, which drove Temporal MutableState lock contention in stage
// (4h window: ~4.7k busyworkflow errors / minute).
//
// The signal model here replaces the timer with a single "wake-up"
// signal that external writers send whenever they mutate state the
// workflow cares about. The workflow re-reads authoritative state when
// the signal fires (same logic as the old per-tick read). History
// growth becomes O(state-transitions) instead of O(wall-clock-seconds).
package processjobsignals

import "fmt"

// SignalName is the single wake-up signal sent to ProcessJob. Payload
// carries a reason string purely for observability; the workflow does
// not branch on it — it re-fetches all state on any wake-up.
const SignalName = "processjob_wakeup"

// Reasons for wake-up signals. Purely informational; workflow treats
// all reasons identically.
const (
	ReasonJobStatusChanged      = "job_status_changed"
	ReasonJobExecutionCreated   = "job_execution_created"
	ReasonJobExecutionStatusChg = "job_execution_status_changed"
	ReasonRunnerStatusChanged   = "runner_status_changed"
	ReasonRunnerRestarted       = "runner_restarted"
)

// WakeUp is the signal payload.
type WakeUp struct {
	// Reason is one of the Reason* constants. Logged by the workflow.
	Reason string `json:"reason"`
}

// WorkflowID returns the deterministic Temporal workflow ID used by the
// ProcessJob workflow for the given runner-job. Writers that need to
// signal ProcessJob use this helper so senders and spawners agree.
//
// There is exactly one ProcessJob workflow per runner-job, so colliding
// IDs are not a concern.
func WorkflowID(jobID string) string {
	return fmt.Sprintf("processjob-%s", jobID)
}
