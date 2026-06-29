package interests

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Matches reports whether the given Interests config wants to receive this
// lifecycle event. Hooks (Slack channel sub, webhook) call this once per
// subscription before dispatching.
//
// Semantics:
//   - AllEvents=true: matches every classifiable event.
//   - Empty Resources (and AllEvents=false): never matches.
//   - Resource missing from Resources: never matches.
//   - Resource present:
//   - Approval events are gated by ApprovalRequests / ApprovalResponses
//     (independent of Outcome and Ops).
//   - Drift-detected events are gated by DriftDetected (independent of
//     Outcome and Ops).
//   - Lifecycle events:
//   - Ops empty → every sub-op for this resource matches.
//   - Ops non-empty → only the listed sub-op matches.
//     The Ops filter applies ONLY to lifecycle events; it never silences
//     approval or drift-detected events.
//   - Outcome:
//   - "none" → no lifecycle events (drift / approval still flow if enabled).
//   - "" / "all" → every started + terminal event.
//   - "completion" → terminal events only (suppress started).
//   - "failures" → only failed/cancelled terminal events.
//
// outcome may be nil at BeforePhase (started). db may be nil — when nil,
// classification falls back to the WorkflowType-only path which is enough
// for execute-workflow events but skips step-scoped enrichment.
func Matches(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome, db *gorm.DB, in Interests) bool {
	if !in.AllEvents && len(in.Resources) == 0 {
		return false
	}

	f := classify(event, outcome, db)
	if !f.Resolved {
		return false
	}

	// Drift workflows (drift_run, drift_run_reprovision_sandbox) emit a flood
	// of started/completed lifecycle events on every cron tick — including
	// clean scans where nothing drifted. That noise is never useful: drift is
	// signaled exclusively through the dedicated drift-detected event class
	// (eventClassDriftDetected) which fires only when the plan-only check
	// observes actual changes.
	//
	// We key off event.WorkflowType (the parent workflow envelope), not the
	// classified facts.Op, because steps inside a drift workflow classify
	// independently — a "runner healthy" step inside drift_run_reprovision_sandbox
	// classifies as (runners, reprovision), a pre-reprovision lifecycle action
	// classifies as (actions, run), etc. Filtering on facts.Op alone would
	// only suppress the outer envelope and the single sandbox-plan step,
	// leaking every other step event to subscribers. Suppress the entire
	// lifecycle tree (including under AllEvents) so subscribers opt into
	// drift via the DriftDetected flag, not by listing "drift" in Ops.
	if f.EventClass == eventClassLifecycle &&
		(event.WorkflowType == "drift_run" || event.WorkflowType == "drift_run_reprovision_sandbox") {
		return false
	}

	if in.AllEvents {
		return true
	}

	cfg, ok := in.Resources[f.Resource]
	if !ok {
		return false
	}

	switch f.EventClass {
	case eventClassApprovalRequest:
		return cfg.ApprovalRequests
	case eventClassApprovalResponse:
		return cfg.ApprovalResponses
	case eventClassDriftDetected:
		return cfg.DriftDetected
	case eventClassRoleChange:
		return cfg.RoleChanges
	case eventClassInputsUpdated:
		return cfg.InputsUpdated
	case eventClassConfigSynced:
		return cfg.ConfigSynced
	case eventClassLifecycle:
		if len(cfg.Ops) > 0 && !contains(cfg.Ops, f.Op) {
			return false
		}
		// Lifecycle outcome filter. Empty Outcome is treated as OutcomeAll
		// for forward-compatibility with rows persisted before this field
		// existed.
		switch cfg.Outcome {
		case OutcomeNone:
			return false
		case OutcomeFailures:
			return f.IsFailureOrCancellation()
		case OutcomeCompletion:
			return f.IsTerminal()
		default:
			// OutcomeAll / "": every started + terminal event matches.
			return true
		}
	}
	return false
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
