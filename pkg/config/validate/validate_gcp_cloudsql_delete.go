package validate

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
)

var gcpCloudSQLDeletePermissions = map[string]struct{}{
	"cloudsql.instances.delete":  {},
	"cloudsql.databases.delete":  {},
	"cloudsql.backupRuns.delete": {},
}

var gcpCloudSQLDeleteRoles = map[string]struct{}{
	"roles/admin":              {},
	"roles/owner":              {},
	"roles/writer":             {},
	"roles/editor":             {},
	"roles/cloudsql.admin":     {},
	"roles/iam.databasesAdmin": {},
}

func ValidateGCPCloudSQLDeleteOutsideBreakGlass(a *config.AppConfig) error {
	if a.Permissions == nil || a.BreakGlass == nil {
		return nil
	}

	breakGlassRole := ""
	for _, role := range a.BreakGlass.Roles {
		if role != nil && isGCP(role) && grantsCloudSQLDeletePermission(role) {
			breakGlassRole = role.Name
			break
		}
	}
	if breakGlassRole == "" {
		return nil
	}

	var warnings []string
	for _, role := range permissionRoles(a.Permissions) {
		if !isGCP(role) {
			continue
		}
		for _, grant := range cloudSQLDeleteGrants(role) {
			warnings = append(warnings, fmt.Sprintf(
				"[permissions:%s] %s grants Cloud SQL delete, which this app reserves for break_glass role %q",
				role.Name, grant, breakGlassRole,
			))
		}
	}

	if len(warnings) == 0 {
		return nil
	}

	return config.ErrConfig{
		Description: strings.Join(warnings, "\n"),
		Warning:     true,
	}
}

func cloudSQLDeleteGrants(role *config.AppAWSIAMRole) []string {
	var grants []string
	for _, policy := range role.Policies {
		if _, ok := gcpCloudSQLDeleteRoles[policy.GCPPredefinedRole]; ok {
			grants = append(grants, fmt.Sprintf("policy %q (%s)", policy.Name, policy.GCPPredefinedRole))
		}
		for _, perm := range policy.GCPPermissions {
			if _, ok := gcpCloudSQLDeletePermissions[perm]; ok {
				grants = append(grants, fmt.Sprintf("policy %q (%s)", policy.Name, perm))
			}
		}
	}
	return grants
}

func grantsCloudSQLDeletePermission(role *config.AppAWSIAMRole) bool {
	for _, policy := range role.Policies {
		for _, perm := range policy.GCPPermissions {
			if _, ok := gcpCloudSQLDeletePermissions[perm]; ok {
				return true
			}
		}
	}
	return false
}

func isGCP(role *config.AppAWSIAMRole) bool {
	return strings.EqualFold(role.CloudPlatform, "gcp")
}

func permissionRoles(p *config.PermissionsConfig) []*config.AppAWSIAMRole {
	candidates := []*config.AppAWSIAMRole{p.ProvisionRole, p.MaintenanceRole, p.DeprovisionRole}
	candidates = append(candidates, p.CustomRoles...)
	candidates = append(candidates, p.Roles...)

	seen := make(map[*config.AppAWSIAMRole]struct{}, len(candidates))
	out := make([]*config.AppAWSIAMRole, 0, len(candidates))
	for _, role := range candidates {
		if role == nil {
			continue
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}
