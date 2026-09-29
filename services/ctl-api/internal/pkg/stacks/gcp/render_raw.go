package gcp

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GCPRoleRaw struct {
	Name            string
	Permissions     []string
	Policies        map[string][]string
	PredefinedRole  string
	PredefinedRoles []string
}

type GCPOpRoleRaw struct {
	Permissions     []string
	Policies        map[string][]string
	PredefinedRole  string
	PredefinedRoles []string
}

func extractRolePermissions(role app.AppAWSIAMRoleConfig) []string {
	var perms []string
	for _, policy := range role.Policies {
		perms = append(perms, policy.GCPPermissions...)
	}
	return perms
}

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

func legacyPredefinedRole(roles []string) string {
	if len(roles) == 0 {
		return ""
	}
	return roles[len(roles)-1]
}
