package operationroles

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

// DefaultRoleForWorkflowType returns the default role name to use for a step
// based on the type of the parent install workflow it runs within.
//
// Provisioning workflows run every step as the provision role, deprovisioning
// workflows run every step as the deprovision role, and everything else (manual
// deploys, input updates, action runs, drift, etc.) uses the maintenance role.
//
// This is only the lowest-precedence default: runtime, break-glass, entity and
// matrix role selections still take precedence over it.
func DefaultRoleForWorkflowType(appCfg *app.AppConfig, workflowType app.WorkflowType) string {
	switch workflowType {
	case app.WorkflowTypeProvision,
		app.WorkflowTypeReprovision,
		app.WorkflowTypeReprovisionSandbox,
		app.WorkflowTypeDriftRunReprovisionSandbox:
		return appCfg.PermissionsConfig.ProvisionRole.Name
	case app.WorkflowTypeDeprovision,
		app.WorkflowTypeDeprovisionSandbox:
		return appCfg.PermissionsConfig.DeprovisionRole.Name
	default:
		return appCfg.PermissionsConfig.MaintenanceRole.Name
	}
}

// defaultRoleForWorkflow returns the lowest-precedence default role for a step.
// A nil workflow means workflow-type defaulting does not apply (the config flag
// is off, or the parent workflow could not be resolved), so it falls back to the
// maintenance role.
func defaultRoleForWorkflow(appCfg *app.AppConfig, flw *app.Workflow) string {
	if flw == nil {
		return appCfg.PermissionsConfig.MaintenanceRole.Name
	}
	return DefaultRoleForWorkflowType(appCfg, flw.Type)
}
