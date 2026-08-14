// Package build converts parsed app config into ctl-api database models. Both
// the HTTP handlers and the branch-sync DB syncer call it, so the two sync paths
// cannot drift. Builders are pure; the caller resolves anything needing a
// database.
package build

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/azureroles"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func IAMPolicies(policies []config.AppAWSIAMPolicy, appConfigID string) []app.AppAWSIAMPolicyConfig {
	out := make([]app.AppAWSIAMPolicyConfig, 0, len(policies))
	for _, policy := range policies {
		out = append(out, app.AppAWSIAMPolicyConfig{
			AppConfigID:       appConfigID,
			ManagedPolicyName: policy.ManagedPolicyName,
			Name:              policy.Name,
			Contents:          dbgenerics.ToJSON(policy.Contents),
			GCPPermissions:    policy.GCPPermissions,
			GCPPredefinedRole: policy.GCPPredefinedRole,
			AzureActions:      policy.AzureActions,
			AzureBuiltInRoles: policy.AzureBuiltInRoles,
		})
	}
	return out
}

func IAMRole(role *config.AppAWSIAMRole, appConfigID string, roleType app.AWSIAMRoleType) app.AppAWSIAMRoleConfig {
	return app.AppAWSIAMRoleConfig{
		AppConfigID:             appConfigID,
		CloudPlatform:           role.CloudPlatform,
		Type:                    roleType,
		Name:                    role.Name,
		Description:             role.Description,
		DisplayName:             role.DisplayName,
		PermissionsBoundaryJSON: dbgenerics.ToJSON(role.PermissionsBoundary),
		EnabledInStack:          generics.NewNullBoolFromPtr(role.EnabledInStack),
		Policies:                IAMPolicies(role.Policies, appConfigID),
	}
}

func IAMRoles(roles []*config.AppAWSIAMRole, appConfigID string, roleType app.AWSIAMRoleType) []app.AppAWSIAMRoleConfig {
	out := make([]app.AppAWSIAMRoleConfig, 0, len(roles))
	for _, role := range roles {
		if role == nil {
			continue
		}
		out = append(out, IAMRole(role, appConfigID, roleType))
	}
	return out
}

// ValidatePolicyMutualExclusivity rejects two grant mechanisms for one cloud:
// the renderers would silently pick one and under-permission the role.
func ValidatePolicyMutualExclusivity(roleName string, policies []config.AppAWSIAMPolicy) error {
	for _, p := range policies {
		if p.Contents != "" && p.ManagedPolicyName != "" {
			return fmt.Errorf("role %q policy %q: contents and managed_policy_name are mutually exclusive; specify one or the other", roleName, p.Name)
		}
		if len(p.GCPPermissions) > 0 && p.GCPPredefinedRole != "" {
			return fmt.Errorf("role %q policy %q: gcp_permissions and gcp_predefined_role are mutually exclusive; use gcp_permissions for fine-grained custom permissions or gcp_predefined_role for a Google-managed role, not both", roleName, p.Name)
		}
		if len(p.AzureActions) > 0 && len(p.AzureBuiltInRoles) > 0 {
			return fmt.Errorf("role %q policy %q: azure_actions and azure_built_in_roles are mutually exclusive; use azure_actions for a fine-grained custom role or azure_built_in_roles for Azure-managed roles, not both", roleName, p.Name)
		}
	}
	return nil
}

// ValidateAzureBuiltInRoles rejects a built-in role that cannot be resolved to a
// definition GUID. ARM assignments reference a definition by GUID with no name
// lookup, and the renderer forwards an unresolvable value verbatim -- so a typo,
// or a real role absent from the name map, passes sync and generation and instead
// fails the customer's stack deployment with InvalidRoleDefinitionId. Catch it
// here, where the error can name the offending value.
func ValidateAzureBuiltInRoles(roleName string, policies []config.AppAWSIAMPolicy) error {
	for _, p := range policies {
		for _, r := range p.AzureBuiltInRoles {
			if azureroles.Resolvable(r) {
				continue
			}
			return fmt.Errorf(
				"role %q policy %q: azure_built_in_roles entry %q is neither a role definition GUID nor a known role name; pass the GUID or one of: %s",
				roleName, p.Name, r, strings.Join(azureroles.KnownNames(), ", "),
			)
		}
	}
	return nil
}

// ValidateInlinePolicyContents rejects malformed inline policies at sync; the
// AWS Terraform path would otherwise render them empty and under-permission the
// runner at apply time.
func ValidateInlinePolicyContents(roles []app.AppAWSIAMRoleConfig) error {
	for _, role := range roles {
		for _, policy := range role.Policies {
			if len(policy.Contents) == 0 || policy.ManagedPolicyName != "" {
				continue
			}
			var doc struct {
				Statement []map[string]json.RawMessage `json:"Statement"`
			}
			if err := json.Unmarshal(policy.Contents, &doc); err != nil {
				return fmt.Errorf("role %q policy %q: inline policy contents must be a JSON IAM policy document: %w", role.Name, policy.Name, err)
			}
			// An empty Statement list is allowed: Azure configs use it as a
			// placeholder because the real grants come from RBAC.
			for i, stmt := range doc.Statement {
				if _, ok := stmt["Effect"]; !ok {
					return fmt.Errorf("role %q policy %q: Statement[%d] missing required Effect field", role.Name, policy.Name, i)
				}
			}
		}
	}
	return nil
}
