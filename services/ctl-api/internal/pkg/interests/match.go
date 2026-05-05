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
//   - Ops empty → every sub-op for this resource matches.
//   - Ops non-empty → only the listed sub-op matches.
//   - Approval events are gated by ApprovalRequests / ApprovalResponses
//     (independent of Outcome).
//   - Lifecycle events apply Outcome:
//   - "" / "all" → every started + terminal event.
//   - "completion" → terminal events only (suppress started).
//   - "failures" → only failed/cancelled terminal events.
//
// outcome may be nil at BeforePhase (started). db may be nil — when nil,
// classification falls back to the WorkflowType-only path which is enough
// for execute-workflow events but skips step-scoped enrichment.
func Matches(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome, db *gorm.DB, in Interests) bool {
	if in.AllEvents {
		return true
	}
	if len(in.Resources) == 0 {
		return false
	}

	f := classify(event, outcome, db)
	if !f.Resolved {
		return false
	}

	cfg, ok := in.Resources[f.Resource]
	if !ok {
		return false
	}

	if len(cfg.Ops) > 0 && !contains(cfg.Ops, f.Op) {
		return false
	}

	switch f.EventClass {
	case eventClassApprovalRequest:
		return cfg.ApprovalRequests
	case eventClassApprovalResponse:
		return cfg.ApprovalResponses
	case eventClassDriftDetected:
		return cfg.DriftDetected
	}

	// Lifecycle outcome filter. Empty Outcome is treated as OutcomeAll for
	// forward-compatibility with rows persisted before this field existed.
	switch cfg.Outcome {
	case OutcomeFailures:
		return f.IsFailureOrCancellation()
	case OutcomeCompletion:
		return f.IsTerminal()
	default:
		// OutcomeAll / "": every started + terminal event matches.
		return true
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
