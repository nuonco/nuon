package interests

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// signalType* mirrors the SignalType strings emitted by the public lifecycle
// primitives. Duplicated as locals (instead of importing the
// installs/signals/v2 tree) to keep this package free of domain imports —
// otherwise the slack/webhook hook models would create an import cycle once
// they embed an Interests column.
const (
	signalTypeExecuteWorkflow              signal.SignalType = "execute-workflow"
	signalTypeExecuteWorkflowStep          signal.SignalType = "execute-workflow-step"
	signalTypeWorkflowStepApprovalRequest  signal.SignalType = "workflow-step-approval-request"
	signalTypeWorkflowStepApprovalResponse signal.SignalType = "workflow-step-approval-response"
	signalTypeDriftDetected                signal.SignalType = "drift-detected"
	// signalTypeOnInactive mirrors the runners/signals/oninactive
	// SignalType. The signal fires once per runner-process when its heartbeat
	// has been quiet long enough (5 min) for processhealthcheck to give up on
	// it; the classifier maps it onto (runners, inactive) so subscribers can
	// opt into a notification.
	signalTypeOnInactive      signal.SignalType = "on_inactive"
	signalTypeStackRun        signal.SignalType = "stack-run"
	signalTypeRoleChange      signal.SignalType = "role-change"
	signalTypeInputsUpdated   signal.SignalType = "inputs-updated"
	signalTypeAppConfigSynced signal.SignalType = "app-config-synced"
)

// stepTargetType* mirror the WorkflowStepTargetType strings declared in
// services/ctl-api/internal/app/workflow_step.go. Duplicated as locals for
// the same import-cycle reason as signalType*.
const (
	stepTargetInstallSandboxRun        = "install_sandbox_run"
	stepTargetInstallSandboxRuns       = "install_sandbox_runs"
	stepTargetInstallDeploy            = "install_deploy"
	stepTargetInstallDeploys           = "install_deploys"
	stepTargetInstallActionWorkflowRun = "install_action_workflow_run"
	stepTargetInstallActionRuns        = "install_action_workflow_runs"
	stepTargetInstallRunnerUpdate      = "install_runner_update"
	stepTargetInstallCloudFormation    = "install_cloudformation_stack"
	stepTargetInstallStackVersions     = "install_stack_versions"
	stepTargetRunners                  = "runners"
)

// eventClass disambiguates the three top-level event flavours produced by
// the queue dispatcher: workflow/step lifecycle (started/succeeded/failed/
// cancelled), approval handshake request, and approval handshake response.
type eventClass int

const (
	eventClassUnknown eventClass = iota
	eventClassLifecycle
	eventClassApprovalRequest
	eventClassApprovalResponse
	eventClassDriftDetected
	eventClassRoleChange
	eventClassInputsUpdated
	eventClassConfigSynced
)

// approvalResponseType is the resolved approved/rejected outcome of an
// approval-response signal. Empty string means the classifier could not
// resolve it (no DB, lookup error, or no responses on the row yet).
type approvalResponseType string

const (
	approvalResponseApproved approvalResponseType = "approved"
	approvalResponseRejected approvalResponseType = "rejected"
)

// facts is the internal classification result. Public Classify / Matches
// build their slugs / decisions exclusively from this struct so the two
// public entry points share semantics.
type facts struct {
	// Resolved is true when classification produced a usable
	// (resource, op, eventClass) triple. False means the event isn't one
	// the picker filters; the matcher returns false in that case.
	Resolved bool

	Resource   ResourceKind
	Op         string
	EventClass eventClass

	Phase   signal.SignalPhase
	Outcome *signal.SignalPhaseOutcome

	// ApprovalResponse is set for approval-response events when the
	// approved/rejected outcome could be resolved (DB lookup). Empty
	// otherwise; the response slug falls back to the generic form.
	ApprovalResponse approvalResponseType
}

// IsTerminal reports whether this is an after-phase / final event (vs the
// before-phase "started" emission).
func (f facts) IsTerminal() bool {
	return f.Outcome != nil
}

// IsFailureOrCancellation reports whether the terminal status is
// failed or cancelled. Always false at started time.
func (f facts) IsFailureOrCancellation() bool {
	if f.Outcome == nil {
		return false
	}
	return f.Outcome.Status == signal.SignalStatusError ||
		f.Outcome.Status == signal.SignalStatusCancelled
}

// Classify returns the slug list for a single event. The webhook hook stamps
// these onto the outbound payload's top-level `interests` array so consumers
// can route by prefix without re-implementing the classifier.
//
// outcome may be nil (BeforePhase / "started" emission). db may be nil — when
// nil, classification falls back to the WorkflowType-only path and skips
// step / approval-response disambiguation.
func Classify(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome, db *gorm.DB) []string {
	f := classify(event, outcome, db)
	if !f.Resolved {
		return nil
	}
	return slugsForFacts(f)
}

// classify is the internal entry point shared by Classify and Matches.
// It is split from the public APIs so the matcher doesn't pay the cost of
// allocating a slug slice when the only output it needs is a boolean.
func classify(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome, db *gorm.DB) facts {
	// Backfill WorkflowType from the install_workflows table when the event
	// carries a workflow id but the in-payload workflow type is empty. This
	// happens for queue signals enqueued before the WorkflowType field was
	// added to the executeworkflowstepgroup / executeworkflowstep payloads —
	// their persisted JSON has workflow_type=""; without backfill the drift
	// suppression and step-resolution paths can't tell what kind of workflow
	// they belong to. Cheap (one indexed PK lookup) and only fires on the
	// legacy/in-flight code path; new signals carry the field inline.
	if event.WorkflowType == "" && event.WorkflowID != "" && db != nil {
		var row struct{ Type string }
		if err := db.WithContext(context.Background()).
			Table("install_workflows").
			Select("type").
			Where("id = ?", event.WorkflowID).
			Scan(&row).Error; err == nil && row.Type != "" {
			event.WorkflowType = row.Type
		}
	}

	f := facts{
		Phase:   event.Phase,
		Outcome: outcome,
	}

	switch event.SignalType {
	case signalTypeExecuteWorkflow:
		res, op, ok := workflowResolution(event.WorkflowType)
		if !ok {
			return f
		}
		f.Resource = res
		f.Op = op
		f.EventClass = eventClassLifecycle
		f.Resolved = true
		return f

	case signalTypeExecuteWorkflowStep:
		stepTarget := lookupStepTargetType(event, db)
		res, op, ok := stepResolution(stepTarget, event.WorkflowType)
		if !ok {
			return f
		}
		f.Resource = res
		f.Op = op
		f.EventClass = eventClassLifecycle
		f.Resolved = true
		return f

	case signalTypeWorkflowStepApprovalRequest:
		stepTarget := lookupStepTargetType(event, db)
		res, op, ok := stepResolution(stepTarget, event.WorkflowType)
		if !ok {
			return f
		}
		f.Resource = res
		f.Op = op
		f.EventClass = eventClassApprovalRequest
		f.Resolved = true
		return f

	case signalTypeWorkflowStepApprovalResponse:
		stepTarget := lookupStepTargetType(event, db)
		res, op, ok := stepResolution(stepTarget, event.WorkflowType)
		if !ok {
			return f
		}
		f.Resource = res
		f.Op = op
		f.EventClass = eventClassApprovalResponse
		f.ApprovalResponse = lookupApprovalResponse(event, db)
		f.Resolved = true
		return f

	case signalTypeOnInactive:
		// oninactive is enqueued by processhealthcheck once a runner-process
		// has been silent for 5 minutes — there's no parent workflow and no
		// step target, so we don't go through stepResolution; we just emit
		// the (runners, inactive) lifecycle classification directly.
		f.Resource = ResourceRunners
		f.Op = "inactive"
		f.EventClass = eventClassLifecycle
		f.Resolved = true
		return f

	case signalTypeDriftDetected:
		// Drift-detected events fire from inside the plan-only check of a
		// drift_run / drift_run_reprovision_sandbox workflow. The step's
		// target type (install_deploys / install_sandbox_runs) plus the
		// parent workflow type land us on (components, drift) or
		// (sandboxes, drift). The fallback paths in stepResolution and
		// stepResolutionFromParent cover the no-DB / unknown-target case
		// for these two parent workflow types.
		stepTarget := lookupStepTargetType(event, db)
		res, op, ok := stepResolution(stepTarget, event.WorkflowType)
		if !ok {
			return f
		}
		f.Resource = res
		f.Op = op
		f.EventClass = eventClassDriftDetected
		f.Resolved = true
		return f

	case signalTypeStackRun:
		f.Resource = ResourceStacks
		f.Op = "stack_run"
		f.EventClass = eventClassLifecycle
		f.Resolved = true
		return f

	case signalTypeRoleChange:
		f.Resource = ResourceStacks
		f.Op = "role_change"
		f.EventClass = eventClassRoleChange
		f.Resolved = true
		return f

	case signalTypeInputsUpdated:
		f.Resource = ResourceStacks
		f.Op = "inputs_updated"
		f.EventClass = eventClassInputsUpdated
		f.Resolved = true
		return f

	case signalTypeAppConfigSynced:
		f.Resource = ResourceAppBranches
		f.Op = "run"
		f.EventClass = eventClassConfigSynced
		f.Resolved = true
		return f
	}

	return f
}

// workflowResolution maps a WorkflowType to (resource, sub-op) for top-level
// (execute-workflow) events. Returns ok=false for unknown / not-yet-modelled
// workflow types (e.g. app_branches_*, app_config_build, action_workflow_run
// is intentionally excluded from this top-level path because a parent
// workflow can also be an action_workflow_run).
//
// `manual_deploy` and `deploy_components` envelope workflows are intentionally
// not classified — there is no first-class "deploy run" resource. Per-component
// step events inside those workflows still classify as `(components, deploy)`,
// and `all_events: true` subscribers still receive the envelope event in the
// raw payload.
func workflowResolution(wfType string) (ResourceKind, string, bool) {
	switch wfType {
	// installs.*
	case "provision":
		return ResourceInstalls, "provision", true
	case "deprovision":
		return ResourceInstalls, "deprovision", true
	case "reprovision":
		return ResourceInstalls, "reprovision", true

	// install_configurations.*
	case "input_update":
		return ResourceInstallConfigurations, "inputs", true
	case "sync_secrets":
		return ResourceInstallConfigurations, "secrets", true

	// components.*
	case "teardown_component", "teardown_components":
		return ResourceComponents, "teardown", true
	case "drift_run":
		return ResourceComponents, "drift", true

	// sandboxes.*
	case "reprovision_sandbox":
		return ResourceSandboxes, "reprovision", true
	case "drift_run_reprovision_sandbox":
		return ResourceSandboxes, "drift", true
	case "deprovision_sandbox":
		return ResourceSandboxes, "deprovision", true

	// actions.*
	case "action_workflow_run":
		return ResourceActions, "run", true

	// installs.* (runbook orchestration)
	case "runbook_run":
		return ResourceInstalls, "runbook", true

	// app_branches.*
	case "app_branches_manual_update",
		"app_branches_config_repo_update",
		"app_branches_component_repo_update":
		return ResourceAppBranches, "run", true
	}

	return "", "", false
}

// stepResolution maps a step target type (and the parent WorkflowType) onto
// (resource, sub-op) for execute-workflow-step and approval events. Returns
// ok=false when the step target isn't part of the modelled taxonomy.
//
// Step events do not project up to the deploys.* sub-ops; a per-component
// step inside a "deploy_components" workflow is classified as
// (components, deploy), not (deploys, components). The deploys.* slugs only
// apply at the outer workflow level.
//
// When stepTargetType is empty (lookup failed or no DB available), the
// resolver falls back to a heuristic based on the parent WorkflowType. This
// keeps approval events classifiable in unit tests / DB-less callers and is
// good enough for the common case where a workflow's steps share a single
// target kind (e.g. every step in a deploy_components workflow is a deploy).
func stepResolution(stepTargetType, parentWorkflowType string) (ResourceKind, string, bool) {
	if stepTargetType == "" {
		return stepResolutionFromParent(parentWorkflowType)
	}
	switch stepTargetType {
	case stepTargetInstallSandboxRun, stepTargetInstallSandboxRuns:
		switch parentWorkflowType {
		case "provision":
			return ResourceSandboxes, "provision", true
		case "drift_run_reprovision_sandbox", "drift_run":
			return ResourceSandboxes, "drift", true
		case "reprovision", "reprovision_sandbox":
			return ResourceSandboxes, "reprovision", true
		case "deprovision", "deprovision_sandbox":
			return ResourceSandboxes, "deprovision", true
		}
		// Fallback: assume provision when parent context is unavailable.
		return ResourceSandboxes, "provision", true

	case stepTargetInstallDeploy, stepTargetInstallDeploys:
		switch parentWorkflowType {
		case "teardown_component", "teardown_components":
			return ResourceComponents, "teardown", true
		case "drift_run":
			return ResourceComponents, "drift", true
		}
		return ResourceComponents, "deploy", true

	case stepTargetInstallActionWorkflowRun, stepTargetInstallActionRuns:
		return ResourceActions, "run", true

	case stepTargetInstallStackVersions:
		// The await-install-stack-version-run step succeeds when the install
		// stack version flips to active — surface that as (stacks, version_active)
		// rather than falling through to the parent provision/reprovision
		// workflow's sandbox-provision classification.
		return ResourceStacks, "version_active", true

	case stepTargetInstallRunnerUpdate, stepTargetInstallCloudFormation, stepTargetRunners:
		switch parentWorkflowType {
		case "reprovision", "reprovision_sandbox", "drift_run_reprovision_sandbox":
			return ResourceRunners, "reprovision", true
		}
		return ResourceRunners, "provision", true
	}

	return "", "", false
}

// stepResolutionFromParent classifies a step (or its approval) using only the
// parent workflow type. Used when the step's target type is unavailable (DB
// lookup failed or DB not provided). Workflows whose steps don't share a
// single target kind — notably "provision", "reprovision", "deprovision" —
// resolve to the most representative kind (sandbox lifecycle), trading a
// little precision for usefulness without DB access.
func stepResolutionFromParent(parentWorkflowType string) (ResourceKind, string, bool) {
	switch parentWorkflowType {
	case "manual_deploy", "deploy_components":
		return ResourceComponents, "deploy", true
	case "drift_run":
		return ResourceComponents, "drift", true
	case "teardown_component", "teardown_components":
		return ResourceComponents, "teardown", true
	case "provision":
		return ResourceSandboxes, "provision", true
	case "reprovision", "reprovision_sandbox":
		return ResourceSandboxes, "reprovision", true
	case "drift_run_reprovision_sandbox":
		return ResourceSandboxes, "drift", true
	case "deprovision", "deprovision_sandbox":
		return ResourceSandboxes, "deprovision", true
	case "action_workflow_run":
		return ResourceActions, "run", true
	case "runbook_run":
		// Runbook steps are predominantly component deploys; the DB-enriched
		// path will refine to the per-step target when available.
		return ResourceComponents, "deploy", true
	case "input_update":
		return ResourceInstallConfigurations, "inputs", true
	case "sync_secrets":
		return ResourceInstallConfigurations, "secrets", true
	case "app_branches_manual_update",
		"app_branches_config_repo_update",
		"app_branches_component_repo_update":
		return ResourceAppBranches, "run", true
	}
	return "", "", false
}

// stepRow mirrors the install_workflow_steps columns we read for step-target
// resolution. Declared locally (instead of importing app.WorkflowStep) to
// avoid an import cycle once the slack/webhook models embed Interests.
type stepRow struct {
	ID             string `gorm:"column:id"`
	StepTargetType string `gorm:"column:step_target_type"`
}

func (stepRow) TableName() string { return "install_workflow_steps" }

// lookupStepTargetType pulls the step's target type from the DB. Returns ""
// when the lookup is impossible (no DB, no StepID) or fails — callers fall
// through to the WorkflowType-only path in stepResolution.
func lookupStepTargetType(event signal.SignalPhaseEvent, db *gorm.DB) string {
	if db == nil || event.StepID == "" {
		return ""
	}
	var row stepRow
	if err := db.WithContext(ctxOrBackground(db)).
		Where(stepRow{ID: event.StepID}).
		First(&row).Error; err != nil {
		return ""
	}
	return row.StepTargetType
}

// approvalResponseRow mirrors install_workflow_step_approval_responses joined
// against install_workflow_step_approvals to surface the most recent response
// type for a given step. Declared locally for the same import-cycle reason.
type approvalResponseRow struct {
	Type string `gorm:"column:type"`
}

// lookupApprovalResponse resolves whether the most recent response on the
// step's approval row was an approve or a reject. Returns "" on any failure
// path; the slug builder falls back to the generic event:approval.response.
func lookupApprovalResponse(event signal.SignalPhaseEvent, db *gorm.DB) approvalResponseType {
	if db == nil || event.StepID == "" {
		return ""
	}
	var row approvalResponseRow
	err := db.WithContext(ctxOrBackground(db)).
		Table("install_workflow_step_approval_responses AS r").
		Select("r.type AS type").
		Joins("JOIN install_workflow_step_approvals AS a ON a.id = r.install_workflow_step_approval_id").
		Where("a.install_workflow_step_id = ?", event.StepID).
		Order("r.created_at DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return ""
	}
	switch approvalResponseType(row.Type) {
	case approvalResponseApproved:
		return approvalResponseApproved
	case approvalResponseRejected:
		return approvalResponseRejected
	}
	return ""
}

// ctxOrBackground returns the gorm session's context if present, falling back
// to context.Background. GORM's *gorm.DB always has Statement.Context unless
// the caller never threaded one through; defensively backstopping is cheap.
func ctxOrBackground(db *gorm.DB) context.Context {
	if db == nil || db.Statement == nil || db.Statement.Context == nil {
		return context.Background()
	}
	return db.Statement.Context
}

// slugsForFacts builds the slug list for a resolved classification.
func slugsForFacts(f facts) []string {
	slugs := []string{
		ResourceSlug(f.Resource),
		OpSlug(f.Resource, f.Op),
	}

	switch f.EventClass {
	case eventClassLifecycle:
		// Lifecycle: started before phase, succeeded/failed/cancelled after.
		if f.Outcome == nil {
			slugs = append(slugs, SlugEventLifecycleStarted)
			return slugs
		}
		switch f.Phase {
		case signal.SignalPhaseCancel:
			slugs = append(slugs, SlugEventLifecycleCancelled, SlugOutcomeCompletion, SlugOutcomeFailures)
			return slugs
		}
		switch f.Outcome.Status {
		case signal.SignalStatusSuccess:
			slugs = append(slugs, SlugEventLifecycleSucceeded, SlugOutcomeCompletion)
		case signal.SignalStatusCancelled:
			slugs = append(slugs, SlugEventLifecycleCancelled, SlugOutcomeCompletion, SlugOutcomeFailures)
		default:
			slugs = append(slugs, SlugEventLifecycleFailed, SlugOutcomeCompletion, SlugOutcomeFailures)
		}
		return slugs

	case eventClassApprovalRequest:
		slugs = append(slugs, SlugEventApprovalRequest)
		return slugs

	case eventClassApprovalResponse:
		slugs = append(slugs, SlugEventApprovalResponse)
		switch f.ApprovalResponse {
		case approvalResponseApproved:
			slugs = append(slugs, SlugEventApprovalResponseApproved)
		case approvalResponseRejected:
			slugs = append(slugs, SlugEventApprovalResponseRejected)
		}
		return slugs

	case eventClassDriftDetected:
		slugs = append(slugs, SlugEventDriftDetected)
		return slugs

	case eventClassRoleChange:
		slugs = append(slugs, SlugEventRoleChange)
		return slugs

	case eventClassInputsUpdated:
		slugs = append(slugs, SlugEventInputsUpdated)
		return slugs

	case eventClassConfigSynced:
		slugs = append(slugs, SlugEventConfigSynced)
		return slugs
	}

	return slugs
}
