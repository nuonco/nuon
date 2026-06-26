package hooks

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// EventTargetsFromEvent resolves the entity ids referenced by a lifecycle
// event into the labels.EventTargets shape consumed by SubscriptionMatch.
// Each id is best-effort and may be empty — Match.matches treats an empty id
// as "no entity of this kind on the event" so a component-only event never
// falsely satisfies an installs filter.
//
// Install resolution mirrors the legacy installIDFromEvent path verbatim
// (event.OwnerType, data.Workflow.OwnerType, then step-derived lookups for
// install_deploys / install_sandbox_runs / install_sandboxes). Component
// and action resolution layer alongside without disturbing it.
//
// This is the package-level implementation shared by the slack and webhook
// signal hooks. Both callers pass their own *gorm.DB; nil db is tolerated
// (lookup steps simply return empty ids, owner-type-derived ids still flow
// through).
func EventTargetsFromEvent(ctx context.Context, db *gorm.DB, event signal.SignalPhaseEvent, data lifecycleEventData) labels.EventTargets {
	t := labels.EventTargets{}

	// Install id ----------------------------------------------------------
	switch {
	case event.OwnerType == "installs" && event.OwnerID != "":
		t.InstallID = event.OwnerID
	case data.Workflow.OwnerType == "installs" && data.Workflow.OwnerID != "":
		t.InstallID = data.Workflow.OwnerID
	}

	// Component id --------------------------------------------------------
	switch {
	case event.OwnerType == "components" && event.OwnerID != "":
		t.ComponentID = event.OwnerID
	case data.Workflow.OwnerType == "components" && data.Workflow.OwnerID != "":
		t.ComponentID = data.Workflow.OwnerID
	}

	// Action id (action_workflows) ----------------------------------------
	switch {
	case event.OwnerType == "action_workflows" && event.OwnerID != "":
		t.ActionID = event.OwnerID
	case data.Workflow.OwnerType == "action_workflows" && data.Workflow.OwnerID != "":
		t.ActionID = data.Workflow.OwnerID
	}

	// App branch id -------------------------------------------------------
	switch {
	case event.OwnerType == "app_branches" && event.OwnerID != "":
		t.AppBranchID = event.OwnerID
	case data.Workflow.OwnerType == "app_branches" && data.Workflow.OwnerID != "":
		t.AppBranchID = data.Workflow.OwnerID
	}

	// Step-derived enrichment. The enrichment in webhook.go has already
	// surfaced ComponentID and SandboxID on data.Step where applicable; we
	// fan out from those plus the step's TargetType to derive install and
	// action ids.
	if data.Step != nil {
		// Step-surfaced component id wins if not already populated.
		if t.ComponentID == "" && data.Step.ComponentID != "" {
			t.ComponentID = data.Step.ComponentID
		}

		switch data.Step.TargetType {
		case string(app.WorkflowStepTargetTypeInstallDeploy),
			string(app.WorkflowStepTargetTypeInstallDeploys):
			if t.InstallID == "" {
				if id := lookupInstallIDFromDeploy(ctx, db, data.Step.TargetID); id != "" {
					t.InstallID = id
				}
			}
		case string(app.WorkflowStepTargetTypeInstallSandboxRun),
			string(app.WorkflowStepTargetTypeInstallSandboxRuns):
			if t.InstallID == "" {
				if id := lookupInstallIDFromSandboxRun(ctx, db, data.Step.TargetID); id != "" {
					t.InstallID = id
				}
			}
		case string(app.WorkflowStepTargetTypeInstallActionWorkflowRun),
			string(app.WorkflowStepTargetTypeInstallActionWorkflowRuns):
			if t.ActionID == "" {
				if id := lookupActionIDFromInstallActionWorkflowRun(ctx, db, data.Step.TargetID); id != "" {
					t.ActionID = id
				}
			}
		case string(app.WorkflowStepTargetTypeInstallStackVersions):
			if t.InstallID == "" {
				if id := lookupInstallIDFromStackVersion(ctx, db, data.Step.TargetID); id != "" {
					t.InstallID = id
				}
			}
		}

		// Sandbox-derived install. The sandbox is owned by exactly one
		// install.
		if t.InstallID == "" && data.Step.SandboxID != "" {
			if id := lookupInstallIDFromSandbox(ctx, db, data.Step.SandboxID); id != "" {
				t.InstallID = id
			}
		}
	}

	return t
}

// lookupInstallIDFromDeploy resolves the install id behind an install_deploys
// row by walking through install_components.install_id. Best-effort: returns
// "" on any DB error or when the row is missing.
func lookupInstallIDFromDeploy(ctx context.Context, db *gorm.DB, deployID string) string {
	if db == nil || deployID == "" {
		return ""
	}
	var row struct {
		InstallID string
	}
	if err := db.WithContext(ctx).
		Table("install_deploys").
		Select("install_components.install_id AS install_id").
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Where("install_deploys.id = ?", deployID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.InstallID
}

// lookupInstallIDFromSandboxRun resolves the install id behind an
// install_sandbox_runs row directly via its install_id column.
func lookupInstallIDFromSandboxRun(ctx context.Context, db *gorm.DB, sandboxRunID string) string {
	if db == nil || sandboxRunID == "" {
		return ""
	}
	var row struct {
		InstallID string
	}
	if err := db.WithContext(ctx).
		Table("install_sandbox_runs").
		Select("install_id").
		Where("id = ?", sandboxRunID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.InstallID
}

// lookupInstallIDFromSandbox resolves the install id behind an
// install_sandboxes row directly via its install_id column.
func lookupInstallIDFromSandbox(ctx context.Context, db *gorm.DB, sandboxID string) string {
	if db == nil || sandboxID == "" {
		return ""
	}
	var row struct {
		InstallID string
	}
	if err := db.WithContext(ctx).
		Table("install_sandboxes").
		Select("install_id").
		Where("id = ?", sandboxID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.InstallID
}

// lookupInstallIDFromStackVersion resolves the install id behind an
// install_stack_versions row directly via its install_id column. The
// await-install-stack-version-run step uses this target type, so install-scoped
// Match (specific installs / label selectors) needs this to fire for the
// (stacks, version_active) event.
func lookupInstallIDFromStackVersion(ctx context.Context, db *gorm.DB, stackVersionID string) string {
	if db == nil || stackVersionID == "" {
		return ""
	}
	var row struct {
		InstallID string
	}
	if err := db.WithContext(ctx).
		Table("install_stack_versions").
		Select("install_id").
		Where("id = ?", stackVersionID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.InstallID
}

// lookupActionIDFromInstallActionWorkflowRun resolves the action_workflow_id
// behind an install_action_workflow_runs row by walking through
// install_action_workflows.action_workflow_id. Best-effort: returns "" on
// any DB error or when the row is unlinked (manual triggers may leave
// install_action_workflow_id null).
func lookupActionIDFromInstallActionWorkflowRun(ctx context.Context, db *gorm.DB, runID string) string {
	if db == nil || runID == "" {
		return ""
	}
	var row struct {
		ActionWorkflowID string
	}
	if err := db.WithContext(ctx).
		Table("install_action_workflow_runs").
		Select("install_action_workflows.action_workflow_id AS action_workflow_id").
		Joins("JOIN install_action_workflows ON install_action_workflows.id = install_action_workflow_runs.install_action_workflow_id").
		Where("install_action_workflow_runs.id = ?", runID).
		Scan(&row).Error; err != nil {
		return ""
	}
	return row.ActionWorkflowID
}
