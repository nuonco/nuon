package authz

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func resolveJobOwner(db *gorm.DB, orgID, jobID string) (permissions.ResolvedTarget, error) {
	var job struct {
		OwnerID, OwnerType, RunnerID string
		Executor                     app.RunnerJobExecutor
	}
	if err := db.Model(&app.RunnerJob{}).Scopes(scopes.WithDisableViews).
		Select("owner_id", "owner_type", "runner_id", "executor").
		Where(app.RunnerJob{ID: jobID, OrgID: orgID}).Take(&job).Error; err != nil {
		return permissions.ResolvedTarget{}, fmt.Errorf("resolve job owner: %w", err)
	}
	if job.Executor != app.RunnerJobExecutorControlPlane && job.Executor != app.RunnerJobExecutorOrgRunner {
		return permissions.ResolvedTarget{}, fmt.Errorf("unsupported job executor %q", job.Executor)
	}
	owner, err := resolveExecutionOwner(db, orgID, job.OwnerType, job.OwnerID)
	if err != nil {
		return permissions.ResolvedTarget{}, err
	}
	if job.RunnerID == "" {
		if job.Executor != app.RunnerJobExecutorControlPlane ||
			(job.OwnerType != "component_builds" && job.OwnerType != "app_sandbox_builds") || owner.InstallID != "" {
			return permissions.ResolvedTarget{}, fmt.Errorf("runner-less job requires a control-plane build owner")
		}
		return owner, nil
	}
	runner, err := resolveTarget(db, permissions.Scope{Kind: permissions.KindRunner, OrgID: orgID, ID: job.RunnerID})
	if err != nil {
		return permissions.ResolvedTarget{}, err
	}
	if owner.InstallID != runner.InstallID || owner.AppID != runner.AppID ||
		(owner.Kind == permissions.KindRunner && owner.ID != runner.ID) {
		return permissions.ResolvedTarget{}, fmt.Errorf("job owner and runner do not match")
	}
	return owner, nil
}

func resolveLogStreamOwner(db *gorm.DB, orgID, streamID string) (permissions.ResolvedTarget, error) {
	var stream struct{ OwnerID, OwnerType string }
	if err := db.Model(&app.LogStream{}).Select("owner_id", "owner_type").
		Where(app.LogStream{ID: streamID, OrgID: orgID}).Take(&stream).Error; err != nil {
		return permissions.ResolvedTarget{}, fmt.Errorf("resolve log stream owner: %w", err)
	}
	if stream.OwnerType == "runner_jobs" {
		return resolveTarget(db, permissions.Scope{Kind: permissions.KindRunnerJob, OrgID: orgID, ID: stream.OwnerID})
	}
	return resolveExecutionOwner(db, orgID, stream.OwnerType, stream.OwnerID)
}

func resolveExecutionOwner(db *gorm.DB, orgID, ownerType, ownerID string) (permissions.ResolvedTarget, error) {
	if ownerID == "" {
		return permissions.ResolvedTarget{}, fmt.Errorf("execution resource requires an owner")
	}
	var model any
	var parentColumn string
	var parentKind permissions.ResourceKind
	switch ownerType {
	case "installs":
		return resolveTarget(db, permissions.Scope{Kind: permissions.KindInstall, OrgID: orgID, ID: ownerID})
	case "apps":
		return resolveTarget(db, permissions.Scope{Kind: permissions.KindApp, OrgID: orgID, ID: ownerID})
	case "runners":
		return resolveTarget(db, permissions.Scope{Kind: permissions.KindRunner, OrgID: orgID, ID: ownerID})
	case "component_builds":
		var build struct{ ComponentConfigConnectionID, ComponentID string }
		if err := db.Model(&app.ComponentBuild{}).Scopes(scopes.WithDisableViews).
			Select("component_config_connection_id", "component_id").
			Where(app.ComponentBuild{ID: ownerID, OrgID: orgID}).Take(&build).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve component build owner: %w", err)
		}
		componentID, err := loadOwnerParentID(db, &app.ComponentConfigConnection{}, orgID, build.ComponentConfigConnectionID, "component_id")
		if err != nil {
			return permissions.ResolvedTarget{}, err
		}
		if build.ComponentID != "" && build.ComponentID != componentID {
			return permissions.ResolvedTarget{}, fmt.Errorf("build and config connection have different components")
		}
		return resolveExecutionOwner(db, orgID, "components", componentID)
	case "components":
		model, parentColumn, parentKind = &app.Component{}, "app_id", permissions.KindApp
	case "app_sandbox_builds":
		model, parentColumn, parentKind = &app.AppSandboxBuild{}, "app_id", permissions.KindApp
	case "app_branch_runs":
		branchID, err := loadOwnerParentID(db, &app.AppBranchRun{}, orgID, ownerID, "app_branch_id")
		if err != nil {
			return permissions.ResolvedTarget{}, err
		}
		return resolveExecutionOwner(db, orgID, "app_branches", branchID)
	case "app_branches":
		model, parentColumn, parentKind = &app.AppBranch{}, "app_id", permissions.KindApp
	case "install_deploys":
		componentID, err := loadOwnerParentID(db, &app.InstallDeploy{}, orgID, ownerID, "install_component_id")
		if err != nil {
			return permissions.ResolvedTarget{}, err
		}
		return resolveExecutionOwner(db, orgID, "install_components", componentID)
	case "install_components":
		var component struct{ InstallID, ComponentID string }
		if err := db.Model(&app.InstallComponent{}).Scopes(scopes.WithDisableViews).Select("install_id", "component_id").
			Where(app.InstallComponent{ID: ownerID, OrgID: orgID}).Take(&component).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve install component owner: %w", err)
		}
		install, err := resolveTarget(db, permissions.Scope{Kind: permissions.KindInstall, OrgID: orgID, ID: component.InstallID})
		if err != nil {
			return permissions.ResolvedTarget{}, err
		}
		parent, err := resolveExecutionOwner(db, orgID, "components", component.ComponentID)
		if err != nil {
			return permissions.ResolvedTarget{}, err
		}
		if parent.AppID != install.AppID {
			return permissions.ResolvedTarget{}, fmt.Errorf("install and component have different app owners")
		}
		return install, nil
	case "install_sandbox_runs", "install_action_workflow_runs":
		var parentModel any = &app.InstallSandbox{}
		model, parentColumn = &app.InstallSandboxRun{}, "install_sandbox_id"
		if ownerType == "install_action_workflow_runs" {
			model, parentModel, parentColumn = &app.InstallActionWorkflowRun{}, &app.InstallActionWorkflow{}, "install_action_workflow_id"
		}
		var run struct {
			InstallID string
			ParentID  *string
		}
		if err := db.Model(model).Scopes(scopes.WithDisableViews).Select("install_id", parentColumn+" AS parent_id").
			Where(struct{ ID, OrgID string }{ownerID, orgID}).Take(&run).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve install run owner: %w", err)
		}
		if run.ParentID != nil {
			parentInstallID, err := loadOwnerParentID(db, parentModel, orgID, *run.ParentID, "install_id")
			if err != nil {
				return permissions.ResolvedTarget{}, err
			}
			if parentInstallID != run.InstallID {
				return permissions.ResolvedTarget{}, fmt.Errorf("run and parent have different install owners")
			}
		}
		return resolveTarget(db, permissions.Scope{Kind: permissions.KindInstall, OrgID: orgID, ID: run.InstallID})
	case "install_sandboxes":
		model, parentColumn, parentKind = &app.InstallSandbox{}, "install_id", permissions.KindInstall
	case "runner_processes":
		model, parentColumn, parentKind = &app.RunnerProcess{}, "runner_id", permissions.KindRunner
	case "runner_operations":
		model, parentColumn, parentKind = &app.RunnerOperation{}, "runner_id", permissions.KindRunner
	case "install_workflow_steps":
		var step struct{ InstallWorkflowID, OwnerID, OwnerType string }
		if err := db.Model(&app.WorkflowStep{}).Scopes(scopes.WithDisableViews).
			Select("install_workflow_id", "owner_id", "owner_type").
			Where(app.WorkflowStep{ID: ownerID, OrgID: orgID}).Take(&step).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve workflow step owner: %w", err)
		}
		if step.InstallWorkflowID == "" {
			return permissions.ResolvedTarget{}, fmt.Errorf("step requires a workflow")
		}
		var workflow struct{ OwnerID, OwnerType string }
		if err := db.Model(&app.Workflow{}).Scopes(scopes.WithDisableViews).Select("owner_id", "owner_type").
			Where(app.Workflow{ID: step.InstallWorkflowID, OrgID: orgID}).Take(&workflow).Error; err != nil {
			return permissions.ResolvedTarget{}, fmt.Errorf("resolve workflow owner: %w", err)
		}
		if step.OwnerID != workflow.OwnerID || step.OwnerType != workflow.OwnerType {
			return permissions.ResolvedTarget{}, fmt.Errorf("step and workflow have different owners")
		}
		if workflow.OwnerType != "installs" && workflow.OwnerType != "apps" && workflow.OwnerType != "app_branches" {
			return permissions.ResolvedTarget{}, fmt.Errorf("unsupported workflow owner type %q", workflow.OwnerType)
		}
		return resolveExecutionOwner(db, orgID, workflow.OwnerType, workflow.OwnerID)
	default:
		return permissions.ResolvedTarget{}, fmt.Errorf("unsupported execution owner type %q", ownerType)
	}
	parentID, err := loadOwnerParentID(db, model, orgID, ownerID, parentColumn)
	if err != nil {
		return permissions.ResolvedTarget{}, err
	}
	return resolveTarget(db, permissions.Scope{Kind: parentKind, OrgID: orgID, ID: parentID})
}

func loadOwnerParentID(db *gorm.DB, model any, orgID, id, column string) (string, error) {
	if id == "" || orgID == "" {
		return "", fmt.Errorf("resource owner requires an ID and tenant")
	}
	var parent struct{ ID string }
	if err := db.Model(model).Scopes(scopes.WithDisableViews).Select(column + " AS id").
		Where(struct{ ID, OrgID string }{id, orgID}).Take(&parent).Error; err != nil {
		return "", fmt.Errorf("resolve %T parent: %w", model, err)
	}
	if parent.ID == "" {
		return "", fmt.Errorf("resource owner has no %s", column)
	}
	return parent.ID, nil
}
