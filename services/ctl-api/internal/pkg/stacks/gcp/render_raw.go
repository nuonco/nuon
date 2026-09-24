package gcp

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GCPRoleRaw is the un-rendered per-role payload the stack SDK consumes for
// break-glass and custom GCP roles.
type GCPRoleRaw struct {
	Name            string
	Permissions     []string
	Policies        map[string][]string
	PredefinedRole  string
	PredefinedRoles []string
}

// GCPOpRoleRaw is the un-rendered payload for a standard operation role.
type GCPOpRoleRaw struct {
	Permissions     []string
	Policies        map[string][]string
	PredefinedRole  string
	PredefinedRoles []string
}

// extractRolePermissions flattens all policy permissions for a role into a
// single list for the raw stack SDK payload.
func extractRolePermissions(role app.AppAWSIAMRoleConfig) []string {
	var perms []string
	for _, policy := range role.Policies {
		perms = append(perms, policy.GCPPermissions...)
	}
	return perms
}

// extractRolePolicyMap returns policy name → permissions (one custom role per
// policy), mirroring extractRolePolicies. Empty policies are skipped; unnamed
// ones fall back to "policy-<index>".
func extractRolePolicyMap(role app.AppAWSIAMRoleConfig) map[string][]string {
	policies := map[string][]string{}
	for i, policy := range role.Policies {
		if len(policy.GCPPermissions) == 0 {
			continue
		}
		name := policy.Name
		if name == "" {
			name = fmt.Sprintf("policy-%d", i)
		}
		policies[name] = policy.GCPPermissions
	}
	if len(policies) == 0 {
		policies = nil
	}
	return policies
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
		perms := extractRolePermissions(role)
		predefined := extractPredefinedRoles(role)
		if len(perms) == 0 && len(predefined) == 0 {
			continue
		}
		raw := GCPOpRoleRaw{
			Permissions:     perms,
			Policies:        extractRolePolicyMap(role),
			PredefinedRole:  legacyPredefinedRole(predefined),
			PredefinedRoles: predefined,
		}
		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provision = raw
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintenance = raw
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovision = raw
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
		perms := extractRolePermissions(role)
		predefined := extractPredefinedRoles(role)
		if len(perms) == 0 && len(predefined) == 0 {
			continue
		}
		out = append(out, GCPRoleRaw{
			Name:            role.Name,
			Permissions:     perms,
			Policies:        extractRolePolicyMap(role),
			PredefinedRole:  legacyPredefinedRole(predefined),
			PredefinedRoles: predefined,
		})
	}
	return out
}

// extractPredefinedRoles returns every predefined role across a role's
// policies, deduplicated, in config order.
func extractPredefinedRoles(role app.AppAWSIAMRoleConfig) []string {
	var roles []string
	seen := map[string]bool{}
	for _, policy := range role.Policies {
		if policy.GCPPredefinedRole == "" || seen[policy.GCPPredefinedRole] {
			continue
		}
		seen[policy.GCPPredefinedRole] = true
		roles = append(roles, policy.GCPPredefinedRole)
	}
	return roles
}

// legacyPredefinedRole is the single role older stack modules bind. It stays
// the last one so existing installs keep the binding they already have and
// only gain the others.
func legacyPredefinedRole(roles []string) string {
	if len(roles) == 0 {
		return ""
	}
	return roles[len(roles)-1]
}
