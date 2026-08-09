package app

import (
	"errors"

	"gorm.io/gorm"
)

// ResolveOwnerAncestry maps a polymorphic (owner_type, owner_id) reference to
// the owning install and/or app — the grantable ancestry stored on child rows
// (install_workflows, runner_jobs, queues, log_streams, terraform_workspaces)
// and used for authorization chains and list filtering.
//
// Both results nil means the resource sits at the org tier (management
// runners, org/VCS queues, unknown or runner-supplied owner types): callers
// must treat that as "org-wide permission required", never as "unowned".
// Unknown owner types and dangling references resolve to the org tier rather
// than erroring, so creation paths never break on ancestry resolution.
//
// Owner types that are themselves denormalized (install_workflows,
// runner_jobs) are read from their own columns, so resolution costs at most
// two point queries and never recurses. Queries deliberately bypass model
// hooks, views, and soft-delete scoping (plain Table reads): ancestry of a
// soft-deleted parent still resolves, matching the backfill.
func ResolveOwnerAncestry(tx *gorm.DB, orgID, ownerType, ownerID string) (installID, appID *string, err error) {
	if ownerID == "" || orgID == "" {
		return nil, nil, nil
	}

	switch ownerType {
	case "orgs", "org":
		return nil, nil, nil
	case "installs", "install":
		return installAncestry(tx, orgID, ownerID)
	case "apps", "app":
		return nil, &ownerID, nil
	case "app_branches":
		return appAncestryVia(tx, orgID, "app_branches", "app_id", ownerID)
	case "components":
		return appAncestryVia(tx, orgID, "components", "app_id", ownerID)
	case "app_sandbox_builds":
		return appAncestryVia(tx, orgID, "app_sandbox_builds", "app_id", ownerID)
	case "component_builds":
		return componentBuildAncestry(tx, orgID, ownerID)
	case "app_branch_runs":
		return appBranchRunAncestry(tx, orgID, ownerID)
	case "install_deploys":
		return installDeployAncestry(tx, orgID, ownerID)
	case "install_sandbox_runs":
		return installAncestryVia(tx, orgID, "install_sandbox_runs", ownerID)
	case "install_action_workflow_runs":
		return installAncestryVia(tx, orgID, "install_action_workflow_runs", ownerID)
	case "install_sandboxes":
		return installAncestryVia(tx, orgID, "install_sandboxes", ownerID)
	case "install_components":
		return installAncestryVia(tx, orgID, "install_components", ownerID)
	case "notebooks":
		return installAncestryVia(tx, orgID, "notebooks", ownerID)
	case "install_workflows":
		return denormalizedAncestry(tx, orgID, "install_workflows", ownerID)
	case "runner_jobs":
		return denormalizedAncestry(tx, orgID, "runner_jobs", ownerID)
	case "install_workflow_steps":
		return workflowStepAncestry(tx, orgID, ownerID)
	case "runners", "runner_operations":
		return runnerAncestry(tx, orgID, ownerID)
	case "runner_groups":
		return runnerGroupAncestry(tx, orgID, ownerID)
	default:
		return nil, nil, nil
	}
}

// installAncestry returns (installID, appID) for an install, or the org tier
// if the install row does not exist.
func installAncestry(tx *gorm.DB, orgID, installID string) (*string, *string, error) {
	var row struct{ AppID string }
	err := tx.Table("installs").
		Select("app_id").
		Where("org_id = ?", orgID).
		Where("id = ?", installID).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	return &installID, &row.AppID, nil
}

// installAncestryVia resolves through a table carrying a direct install_id
// column (install_sandbox_runs, install_components, notebooks, ...).
func installAncestryVia(tx *gorm.DB, orgID, table, id string) (*string, *string, error) {
	var row struct{ InstallID string }
	err := tx.Table(table).
		Select("install_id").
		Where("org_id = ?", orgID).
		Where("id = ?", id).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	return installAncestry(tx, orgID, row.InstallID)
}

// appAncestryVia resolves through a table carrying a direct app column.
func appAncestryVia(tx *gorm.DB, orgID, table, column, id string) (*string, *string, error) {
	var row struct{ AppID string }
	err := tx.Table(table).
		Select(column+" AS app_id").
		Where("org_id = ?", orgID).
		Where("id = ?", id).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	return nil, &row.AppID, nil
}

// denormalizedAncestry reads the ancestry columns of a table that is itself
// denormalized, so owner chains never recurse through polymorphic links.
func denormalizedAncestry(tx *gorm.DB, orgID, table, id string) (*string, *string, error) {
	var row struct {
		InstallID *string
		AppID     *string
	}
	err := tx.Table(table).
		Select("install_id, app_id").
		Where("org_id = ?", orgID).
		Where("id = ?", id).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	return row.InstallID, row.AppID, nil
}

func componentBuildAncestry(tx *gorm.DB, orgID, buildID string) (*string, *string, error) {
	var row struct{ AppID string }
	err := tx.Table("component_builds").
		Select("components.app_id").
		Joins("JOIN component_config_connections ON component_config_connections.id = component_builds.component_config_connection_id").
		Joins("JOIN components ON components.id = component_config_connections.component_id").
		Where("component_builds.org_id = ?", orgID).
		Where("component_builds.id = ?", buildID).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	return nil, &row.AppID, nil
}

func appBranchRunAncestry(tx *gorm.DB, orgID, runID string) (*string, *string, error) {
	var row struct{ AppID string }
	err := tx.Table("app_branch_runs").
		Select("app_branches.app_id").
		Joins("JOIN app_branches ON app_branches.id = app_branch_runs.app_branch_id").
		Where("app_branch_runs.org_id = ?", orgID).
		Where("app_branch_runs.id = ?", runID).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	return nil, &row.AppID, nil
}

func installDeployAncestry(tx *gorm.DB, orgID, deployID string) (*string, *string, error) {
	var row struct{ InstallID string }
	err := tx.Table("install_deploys").
		Select("install_components.install_id").
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Where("install_deploys.org_id = ?", orgID).
		Where("install_deploys.id = ?", deployID).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	return installAncestry(tx, orgID, row.InstallID)
}

func workflowStepAncestry(tx *gorm.DB, orgID, stepID string) (*string, *string, error) {
	var row struct{ InstallWorkflowID string }
	err := tx.Table("install_workflow_steps").
		Select("install_workflow_id").
		Where("org_id = ?", orgID).
		Where("id = ?", stepID).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	return denormalizedAncestry(tx, orgID, "install_workflows", row.InstallWorkflowID)
}

func runnerAncestry(tx *gorm.DB, orgID, runnerID string) (*string, *string, error) {
	var row struct{ RunnerGroupID string }
	err := tx.Table("runners").
		Select("runner_group_id").
		Where("org_id = ?", orgID).
		Where("id = ?", runnerID).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	return runnerGroupAncestry(tx, orgID, row.RunnerGroupID)
}

// runnerGroupAncestry resolves a runner group's owner: install-owned groups
// inherit the install's ancestry, org-owned groups are the org tier.
func runnerGroupAncestry(tx *gorm.DB, orgID, groupID string) (*string, *string, error) {
	var row struct {
		OwnerID   string
		OwnerType string
	}
	err := tx.Table("runner_groups").
		Select("owner_id, owner_type").
		Where("org_id = ?", orgID).
		Where("id = ?", groupID).
		Take(&row).Error
	if err != nil {
		return orgTierOnNotFound(err)
	}
	if row.OwnerType == "installs" || row.OwnerType == "install" {
		return installAncestry(tx, orgID, row.OwnerID)
	}
	return nil, nil, nil
}

func orgTierOnNotFound(err error) (*string, *string, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil
	}
	return nil, nil, err
}
