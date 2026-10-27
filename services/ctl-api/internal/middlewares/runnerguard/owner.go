package runnerguard

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// ResolveInstallID maps a resource owner (owner_type, owner_id) to an install id
// for install-scoped runnerguard checks. It is not Gin middleware.
// owner_type=runner_processes is runner-scoped; use authorizeLogStream instead.
func ResolveInstallID(ctx context.Context, db *gorm.DB, ownerType, ownerID string) (string, error) {
	if ownerType == "" || ownerID == "" {
		return "", fmt.Errorf("owner is required")
	}
	switch ownerType {
	case "installs", "install":
		return ownerID, nil
	case "install_components", "install_sandbox_runs", "install_action_workflow_runs", "install_sandboxes":
		var row struct {
			InstallID string
		}
		err := db.WithContext(ctx).Table(ownerType).Select("install_id").Where("id = ?", ownerID).Take(&row).Error
		if err != nil {
			return "", err
		}
		if row.InstallID == "" {
			return "", fmt.Errorf("owner %s has no install", ownerID)
		}
		return row.InstallID, nil
	case "install_deploys":
		var row struct {
			InstallID string
		}
		err := db.WithContext(ctx).Table("install_deploys").
			Select("install_components.install_id").
			Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
			Where("install_deploys.id = ?", ownerID).
			Take(&row).Error
		if err != nil {
			return "", err
		}
		if row.InstallID == "" {
			return "", fmt.Errorf("deploy %s has no install", ownerID)
		}
		return row.InstallID, nil
	case "install_workflow_steps":
		var row struct {
			InstallID string
		}
		err := db.WithContext(ctx).Table("install_workflow_steps").
			Select("install_workflows.owner_id AS install_id").
			Joins("JOIN install_workflows ON install_workflows.id = install_workflow_steps.install_workflow_id").
			Where("install_workflow_steps.id = ? AND install_workflows.owner_type = ?", ownerID, "installs").
			Take(&row).Error
		if err != nil {
			return "", err
		}
		if row.InstallID == "" {
			return "", fmt.Errorf("workflow step %s has no install", ownerID)
		}
		return row.InstallID, nil
	case "runner_processes":
		// Process streams are authorized against the runner, not an install.
		// Callers must use authorizeLogStream rather than ResolveInstallID alone.
		return "", fmt.Errorf("runner_processes ownership is runner-scoped")
	default:
		return "", fmt.Errorf("unsupported owner type %s", ownerType)
	}
}
