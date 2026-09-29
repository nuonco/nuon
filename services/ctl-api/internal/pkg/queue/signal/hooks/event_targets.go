package hooks

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func EventTargetsFromEvent(ctx context.Context, db *gorm.DB, event signal.SignalPhaseEvent, data lifecycleEventData) labels.EventTargets {
	t := labels.EventTargets{}

	switch {
	case event.OwnerType == "installs" && event.OwnerID != "":
		t.InstallID = event.OwnerID
	case data.Workflow.OwnerType == "installs" && data.Workflow.OwnerID != "":
		t.InstallID = data.Workflow.OwnerID
	case event.InstallID != nil && *event.InstallID != "":
		t.InstallID = *event.InstallID
	}

	switch {
	case event.OwnerType == "components" && event.OwnerID != "":
		t.ComponentID = event.OwnerID
	case data.Workflow.OwnerType == "components" && data.Workflow.OwnerID != "":
		t.ComponentID = data.Workflow.OwnerID
	case event.ComponentID != nil && *event.ComponentID != "":
		t.ComponentID = *event.ComponentID
	}

	switch {
	case event.OwnerType == "action_workflows" && event.OwnerID != "":
		t.ActionID = event.OwnerID
	case data.Workflow.OwnerType == "action_workflows" && data.Workflow.OwnerID != "":
		t.ActionID = data.Workflow.OwnerID
	}

	switch {
	case event.OwnerType == "app_branches" && event.OwnerID != "":
		t.AppBranchID = event.OwnerID
	case data.Workflow.OwnerType == "app_branches" && data.Workflow.OwnerID != "":
		t.AppBranchID = data.Workflow.OwnerID
	}

	if data.Step != nil {
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

		if t.InstallID == "" && data.Step.SandboxID != "" {
			if id := lookupInstallIDFromSandbox(ctx, db, data.Step.SandboxID); id != "" {
				t.InstallID = id
			}
		}
	}

	return t
}

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
