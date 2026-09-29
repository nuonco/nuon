package interests

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func Matches(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome, db *gorm.DB, in Interests) bool {
	if !in.AllEvents && len(in.Resources) == 0 {
		return false
	}

	f := classify(event, outcome, db)
	if !f.Resolved {
		return false
	}

	// why: Drift workflows (drift_run, drift_run_reprovision_sandbox) emit a flood
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
	case eventClassAwaitingRetry:
		if len(cfg.Ops) > 0 && !contains(cfg.Ops, f.Op) {
			return false
		}
		return cfg.Outcome != OutcomeNone
	case eventClassRoleChange:
		return cfg.RoleChanges
	case eventClassInputsUpdated:
		return cfg.InputsUpdated
	case eventClassConfigSynced:
		return cfg.ConfigSynced
	case eventClassComponentUnhealthy, eventClassComponentRecovered:
		return cfg.ComponentHealth
	case eventClassInstallDegraded:
		return cfg.InstallDegraded
	case eventClassRunnerUnhealthy:
		if len(cfg.Ops) > 0 && !contains(cfg.Ops, f.Op) {
			return false
		}
		return cfg.Outcome != OutcomeNone
	case eventClassLifecycle:
		if len(cfg.Ops) > 0 && !contains(cfg.Ops, f.Op) {
			return false
		}
		// why: Lifecycle outcome filter. Empty Outcome is treated as OutcomeAll
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
