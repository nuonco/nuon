package interests

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	signalTypeExecuteWorkflow              signal.SignalType = "execute-workflow"
	signalTypeExecuteWorkflowStep          signal.SignalType = "execute-workflow-step"
	signalTypeWorkflowStepApprovalRequest  signal.SignalType = "workflow-step-approval-request"
	signalTypeWorkflowStepApprovalResponse signal.SignalType = "workflow-step-approval-response"
	signalTypeDriftDetected                signal.SignalType = "drift-detected"
	signalTypeOnInactive                   signal.SignalType = "on_inactive"
	signalTypeRunnerUnhealthy              signal.SignalType = "runner-unhealthy"
	// why: signalTypeWorkflowStepAwaitingRetry mirrors the
	// flow/signals/workflowstepawaitingretry SignalType. The carrier signal
	// itself always succeeds (notification-only), so classification can't
	// use the phase outcome — the event is unconditionally treated as a
	// failure-flavoured, non-terminal "action required" emission.
	signalTypeWorkflowStepAwaitingRetry signal.SignalType = "workflow-step-awaiting-retry"
	signalTypeStackRun                  signal.SignalType = "stack-run"
	signalTypeRoleChange                signal.SignalType = "role-change"
	signalTypeInputsUpdated             signal.SignalType = "inputs-updated"
	signalTypeAppConfigSynced           signal.SignalType = "app-config-synced"
	signalTypeComponentUnhealthy        signal.SignalType = "component-unhealthy"
	signalTypeComponentRecovered        signal.SignalType = "component-recovered"
	signalTypeInstallDegraded           signal.SignalType = "install-degraded"

	signalTypeSyncInstalls      signal.SignalType = "sync-installs"
	signalTypeInstallConfigSync signal.SignalType = "install-config-sync"
	signalTypeLabelAdded        signal.SignalType = "label-added"
	signalTypeAppBranchChanged  signal.SignalType = "app-branch-changed"
)

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

type eventClass int

const (
	eventClassUnknown eventClass = iota
	eventClassLifecycle
	eventClassApprovalRequest
	eventClassApprovalResponse
	eventClassDriftDetected
	eventClassAwaitingRetry
	eventClassRoleChange
	eventClassInputsUpdated
	eventClassConfigSynced
	eventClassComponentUnhealthy
	eventClassComponentRecovered
	eventClassInstallDegraded
	eventClassRunnerUnhealthy
)

type approvalResponseType string

const (
	approvalResponseApproved approvalResponseType = "approved"
	approvalResponseRejected approvalResponseType = "rejected"
)

type facts struct {
	Resolved bool

	Resource   ResourceKind
	Op         string
	EventClass eventClass

	Phase   signal.SignalPhase
	Outcome *signal.SignalPhaseOutcome

	ApprovalResponse approvalResponseType
}

func (f facts) IsTerminal() bool {
	return f.Outcome != nil
}

func (f facts) IsFailureOrCancellation() bool {
	if f.Outcome == nil {
		return false
	}
	return f.Outcome.Status == signal.SignalStatusError ||
		f.Outcome.Status == signal.SignalStatusCancelled
}

func Classify(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome, db *gorm.DB) []string {
	f := classify(event, outcome, db)
	if !f.Resolved {
		return nil
	}
	return slugsForFacts(f)
}

func classify(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome, db *gorm.DB) facts {
	// why: Backfill WorkflowType from the install_workflows table when the event
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
		f.Resource = ResourceRunners
		f.Op = "inactive"
		f.EventClass = eventClassLifecycle
		f.Resolved = true
		return f

	case signalTypeRunnerUnhealthy:
		f.Resource = ResourceRunners
		f.Op = "unhealthy"
		f.EventClass = eventClassRunnerUnhealthy
		f.Resolved = true
		return f

	case signalTypeDriftDetected:
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

	case signalTypeWorkflowStepAwaitingRetry:
		stepTarget := lookupStepTargetType(event, db)
		res, op, ok := stepResolution(stepTarget, event.WorkflowType)
		if !ok {
			return f
		}
		f.Resource = res
		f.Op = op
		f.EventClass = eventClassAwaitingRetry
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

	case signalTypeSyncInstalls:
		f.Resource = ResourceInstallConfigurations
		f.Op = "sync"
		f.EventClass = eventClassLifecycle
		f.Resolved = true
		return f

	case signalTypeInstallConfigSync:
		f.Resource = ResourceInstallConfigurations
		f.Op = "sync"
		f.EventClass = eventClassLifecycle
		f.Resolved = true
		return f

	case signalTypeLabelAdded:
		f.Resource = ResourceInstalls
		f.Op = "label_added"
		f.EventClass = eventClassLifecycle
		f.Resolved = true
		return f

	case signalTypeAppBranchChanged:
		f.Resource = ResourceInstalls
		f.Op = "app_branch_changed"
		f.EventClass = eventClassLifecycle
		f.Resolved = true
		return f

	case signalTypeComponentUnhealthy:
		// why: The component-health evaluator emits this from its own queue, not
		// from a workflow step, so there is no step target to resolve — the
		// (components, health) classification is unconditional. Like "drift",
		// "health" is deliberately absent from SubOps: subscribers opt in via
		// the ComponentHealth flag, never by listing it in Ops.
		f.Resource = ResourceComponents
		f.Op = "health"
		f.EventClass = eventClassComponentUnhealthy
		f.Resolved = true
		return f

	case signalTypeComponentRecovered:
		f.Resource = ResourceComponents
		f.Op = "health"
		f.EventClass = eventClassComponentRecovered
		f.Resolved = true
		return f

	case signalTypeInstallDegraded:
		f.Resource = ResourceInstalls
		f.Op = "health"
		f.EventClass = eventClassInstallDegraded
		f.Resolved = true
		return f
	}

	return f
}

func workflowResolution(wfType string) (ResourceKind, string, bool) {
	switch wfType {
	case "provision":
		return ResourceInstalls, "provision", true
	case "deprovision":
		return ResourceInstalls, "deprovision", true
	case "reprovision":
		return ResourceInstalls, "reprovision", true
	case "reprovision_stack":
		return ResourceInstalls, "reprovision", true

	case "input_update":
		return ResourceInstallConfigurations, "inputs", true
	case "sync_secrets":
		return ResourceInstallConfigurations, "secrets", true

	case "teardown_component", "teardown_components":
		return ResourceComponents, "teardown", true
	case "drift_run":
		return ResourceComponents, "drift", true

	case "reprovision_sandbox":
		return ResourceSandboxes, "reprovision", true
	case "drift_run_reprovision_sandbox":
		return ResourceSandboxes, "drift", true
	case "deprovision_sandbox":
		return ResourceSandboxes, "deprovision", true

	case "action_workflow_run":
		return ResourceActions, "run", true

	case "runbook_run":
		return ResourceInstalls, "runbook", true

	case "app_branches_manual_update",
		"app_branches_config_repo_update",
		"app_branches_component_repo_update":
		return ResourceAppBranches, "run", true
	}

	return "", "", false
}

func stepResolution(stepTargetType, parentWorkflowType string) (ResourceKind, string, bool) {
	switch parentWorkflowType {
	case "app_branches_manual_update",
		"app_branches_config_repo_update",
		"app_branches_component_repo_update":
		return ResourceAppBranches, "run", true
	}

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
		return ResourceStacks, "version_active", true

	case stepTargetInstallRunnerUpdate, stepTargetInstallCloudFormation, stepTargetRunners:
		switch parentWorkflowType {
		case "reprovision", "reprovision_stack", "reprovision_sandbox", "drift_run_reprovision_sandbox":
			return ResourceRunners, "reprovision", true
		}
		return ResourceRunners, "provision", true
	}

	return "", "", false
}

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
	case "reprovision_stack":
		return ResourceRunners, "reprovision", true
	case "drift_run_reprovision_sandbox":
		return ResourceSandboxes, "drift", true
	case "deprovision", "deprovision_sandbox":
		return ResourceSandboxes, "deprovision", true
	case "action_workflow_run":
		return ResourceActions, "run", true
	case "runbook_run":
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

type stepRow struct {
	ID             string `gorm:"column:id"`
	StepTargetType string `gorm:"column:step_target_type"`
}

func (stepRow) TableName() string { return "install_workflow_steps" }

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

type approvalResponseRow struct {
	Type string `gorm:"column:type"`
}

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

func ctxOrBackground(db *gorm.DB) context.Context {
	if db == nil || db.Statement == nil || db.Statement.Context == nil {
		return context.Background()
	}
	return db.Statement.Context
}

func slugsForFacts(f facts) []string {
	slugs := []string{
		ResourceSlug(f.Resource),
		OpSlug(f.Resource, f.Op),
	}

	switch f.EventClass {
	case eventClassLifecycle:
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

	case eventClassAwaitingRetry:
		// why: Projected as a failure: the step failed and parked. Deliberately
		// no outcome:completion slug — the step is not terminal (a real
		// lifecycle event follows once it's retried, skipped, or cancelled).
		slugs = append(slugs, SlugEventLifecycleFailed, SlugOutcomeFailures)
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

	case eventClassComponentUnhealthy:
		// why: No outcome slug: health is not a workflow outcome, and emitting
		// outcome:failures here would silently widen what existing
		// outcome-routed consumers receive. Same choice as drift.detected.
		slugs = append(slugs, SlugEventComponentUnhealthy)
		return slugs

	case eventClassComponentRecovered:
		slugs = append(slugs, SlugEventComponentRecovered)
		return slugs

	case eventClassInstallDegraded:
		slugs = append(slugs, SlugEventInstallDegraded)
		return slugs

	case eventClassRunnerUnhealthy:
		slugs = append(slugs, SlugEventRunnerUnhealthy, SlugOutcomeCompletion, SlugOutcomeFailures)
		return slugs
	}

	return slugs
}
