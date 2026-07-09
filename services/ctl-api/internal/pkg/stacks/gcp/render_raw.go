package gcp

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GCPRoleRaw is the un-rendered per-role payload the stack SDK consumes for
// break-glass and custom GCP roles.
type GCPRoleRaw struct {
	Name           string
	Permissions    []string
	PredefinedRole string
}

// GCPOpRoleRaw is the un-rendered payload for a standard operation role.
type GCPOpRoleRaw struct {
	Permissions    []string
	PredefinedRole string
}

// extractRolePermissions flattens all policy permissions for a role into a
// single list, alongside any predefined role, for the raw stack SDK payload.
func extractRolePermissions(role app.AppAWSIAMRoleConfig) ([]string, string) {
	var perms []string
	var predefinedRole string
	for _, policy := range role.Policies {
		if policy.GCPPredefinedRole != "" {
			predefinedRole = policy.GCPPredefinedRole
		}
		perms = append(perms, policy.GCPPermissions...)
	}
	return perms, predefinedRole
}

// ExtractGCPStandardRolesRaw returns the permissions/predefined-role for the
// standard provision/maintenance/deprovision operation roles.
func ExtractGCPStandardRolesRaw(appCfg *app.AppConfig) (provision, maintenance, deprovision GCPOpRoleRaw) {
	if appCfg == nil {
		return
	}
	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != "gcp" {
			continue
		}
		perms, predefined := extractRolePermissions(role)
		if len(perms) == 0 && predefined == "" {
			continue
		}
		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provision = GCPOpRoleRaw{Permissions: perms, PredefinedRole: predefined}
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintenance = GCPOpRoleRaw{Permissions: perms, PredefinedRole: predefined}
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovision = GCPOpRoleRaw{Permissions: perms, PredefinedRole: predefined}
		}
	}
	return
}

// ExtractGCPRolesRaw returns the raw payload for a list of break-glass or
// custom GCP roles, skipping non-GCP and empty roles.
func ExtractGCPRolesRaw(roles []app.AppAWSIAMRoleConfig) []GCPRoleRaw {
	var out []GCPRoleRaw
	for _, role := range roles {
		if role.CloudPlatform != "gcp" {
			continue
		}
		perms, predefined := extractRolePermissions(role)
		if len(perms) == 0 && predefined == "" {
			continue
		}
		out = append(out, GCPRoleRaw{Name: role.Name, Permissions: perms, PredefinedRole: predefined})
	}
	return out
}
