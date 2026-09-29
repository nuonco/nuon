package operationroles

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

func DefaultRoleForWorkflowType(appCfg *app.AppConfig, workflowType app.WorkflowType) string {
	switch workflowType {
	case app.WorkflowTypeProvision,
		app.WorkflowTypeReprovision,
		app.WorkflowTypeReprovisionStack,
		app.WorkflowTypeReprovisionSandbox,
		app.WorkflowTypeDriftRunReprovisionSandbox,
		app.WorkflowTypeDeployComponents:
		return appCfg.PermissionsConfig.ProvisionRole.Name
	case app.WorkflowTypeDeprovision,
		app.WorkflowTypeDeprovisionSandbox,
		app.WorkflowTypeTeardownComponents:
		return appCfg.PermissionsConfig.DeprovisionRole.Name
	default:
		return appCfg.PermissionsConfig.MaintenanceRole.Name
	}
}

func defaultRoleForWorkflow(appCfg *app.AppConfig, flw *app.Workflow) string {
	if flw == nil {
		return appCfg.PermissionsConfig.MaintenanceRole.Name
	}
	return DefaultRoleForWorkflowType(appCfg, flw.Type)
}
