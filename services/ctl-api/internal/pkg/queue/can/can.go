// Package can provides the canonical "should this long-lived workflow
// continue-as-new because its history is getting big?" check used by every
// looping workflow in the queue subsystem (queue, handler, emitter, sweep).
//
// Use ShouldContinueAsNew from the workflow's await predicate; on true, fall
// through to the workflow's existing continue-as-new path. Time-based
// ceilings (e.g. queue maxAliveTime, emitter cronParentRunDuration,
// sweepMaxAliveTime) are kept where they exist for their own purposes; this
// helper is purely history-aware.
package can

import "go.temporal.io/sdk/workflow"

// HistoryLengthCANThreshold is the workflow history-event count at which a
// long-lived loop should continue-as-new. Temporal's hard fail is 51,200; we
// rotate well below that. Server-side suggestion (GetContinueAsNewSuggested)
// fires earlier (~4K events / ~4MB by default) and is the primary trigger.
const HistoryLengthCANThreshold = 10_000

// ShouldContinueAsNew returns true when this workflow should rotate via
// continue-as-new based on its current history. Combines Temporal's
// server-side suggestion (which covers both event count and byte size) with
// an explicit event-count fallback.
func ShouldContinueAsNew(ctx workflow.Context) bool {
	info := workflow.GetInfo(ctx)
	if info == nil {
		return false
	}
	if info.GetContinueAsNewSuggested() {
		return true
	}
	return info.GetCurrentHistoryLength() >= HistoryLengthCANThreshold
}
