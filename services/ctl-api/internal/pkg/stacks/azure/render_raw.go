package azure

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/azureroles"
)

type AzureRoleRaw struct {
	Name         string
	Actions      []string
	BuiltInRoles []string
}

type AzureOpRoleRaw struct {
	Actions      []string
	BuiltInRoles []string
}

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
