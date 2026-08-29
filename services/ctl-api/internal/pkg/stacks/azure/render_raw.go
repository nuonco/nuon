package azure

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/azureroles"
)

// AzureRoleRaw is the un-rendered per-role payload the stack SDK consumes for
// break-glass and custom Azure roles.
//
// Azure splits a role's grants across two axes, and the Terraform module treats
// them differently: Actions become a subscription-scoped custom role definition,
// BuiltInRoles become direct assignments at resource-group scope.
type AzureRoleRaw struct {
	Name         string
	Actions      []string
	BuiltInRoles []string
}

// AzureOpRoleRaw is the un-rendered payload for a standard operation role.
type AzureOpRoleRaw struct {
	Actions      []string
	BuiltInRoles []string
}

// extractRoleGrants flattens a role's policies into actions and built-in roles.
//
// Built-in roles are resolved to GUIDs here rather than forwarded by name: the
// Terraform module builds a role definition ID from the value verbatim, and
// azureroles.GUID is the only place that knows the name → GUID mapping. The ARM
// path resolves at render time for the same reason.
func extractRoleGrants(role app.AppAWSIAMRoleConfig) ([]string, []string) {
	var actions, builtInRoles []string
	for _, policy := range role.Policies {
		actions = append(actions, policy.AzureActions...)
		for _, r := range policy.AzureBuiltInRoles {
			builtInRoles = append(builtInRoles, azureroles.GUID(r))
		}
	}
	return actions, builtInRoles
}

// ExtractAzureStandardRolesRaw returns the grants for the standard
// provision/maintenance/deprovision operation roles.
func ExtractAzureStandardRolesRaw(appCfg *app.AppConfig) (provision, maintenance, deprovision AzureOpRoleRaw) {
	if appCfg == nil {
		return
	}
	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != string(app.CloudPlatformAzure) {
			continue
		}
		actions, builtIn := extractRoleGrants(role)
		if len(actions) == 0 && len(builtIn) == 0 {
			continue
		}
		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provision = AzureOpRoleRaw{Actions: actions, BuiltInRoles: builtIn}
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintenance = AzureOpRoleRaw{Actions: actions, BuiltInRoles: builtIn}
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovision = AzureOpRoleRaw{Actions: actions, BuiltInRoles: builtIn}
		}
	}
	return
}

// ExtractAzureRolesRaw returns the raw payload for a list of break-glass or
// custom Azure roles, skipping non-Azure and empty roles.
func ExtractAzureRolesRaw(roles []app.AppAWSIAMRoleConfig) []AzureRoleRaw {
	var out []AzureRoleRaw
	for _, role := range roles {
		if role.CloudPlatform != string(app.CloudPlatformAzure) {
			continue
		}
		actions, builtIn := extractRoleGrants(role)
		if len(actions) == 0 && len(builtIn) == 0 {
			continue
		}
		out = append(out, AzureRoleRaw{Name: role.Name, Actions: actions, BuiltInRoles: builtIn})
	}
	return out
}
