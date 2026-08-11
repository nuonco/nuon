package helpers

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// InstallOwnedScope constrains a query over any table with polymorphic
// owner_id/owner_type columns (runner_jobs, terraform_workspaces, …) to the
// rows whose owner resolves to installID. It ORs across every install-bearing
// owner type; a table that never uses a given owner type simply won't match it.
//
// This is the single "does this owner_id/owner_type belong to :install_id?"
// predicate shared by callers that scope polymorphic resources to an install.
func (h *Helpers) InstallOwnedScope(installID string) func(*gorm.DB) *gorm.DB {
	sub := func(model any, where string) *gorm.DB {
		return h.db.Model(model).Select("id").Where(where+" = ?", installID)
	}

	return func(q *gorm.DB) *gorm.DB {
		installComponents := sub(&app.InstallComponent{}, "install_id")
		installSandboxes := sub(&app.InstallSandbox{}, "install_id")
		installSandboxRuns := sub(&app.InstallSandboxRun{}, "install_id")
		installActionRuns := sub(&app.InstallActionWorkflowRun{}, "install_id")
		installDeploys := h.db.Model(&app.InstallDeploy{}).
			Select("install_deploys.id").
			Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
			Where("install_components.install_id = ?", installID)

		return q.Where(
			h.db.Where("owner_type = ? AND owner_id IN (?)", "install_components", installComponents).
				Or("owner_type = ? AND owner_id IN (?)", "install_sandboxes", installSandboxes).
				Or("owner_type = ? AND owner_id IN (?)", "install_sandbox_runs", installSandboxRuns).
				Or("owner_type = ? AND owner_id IN (?)", "install_action_workflow_runs", installActionRuns).
				Or("owner_type = ? AND owner_id IN (?)", "install_deploys", installDeploys),
		)
	}
}
