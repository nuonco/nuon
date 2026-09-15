// Package build converts parsed app config into ctl-api database models. Both
// the HTTP handlers and the branch-sync DB syncer call it, so the two sync paths
// cannot drift. Builders are pure; the caller resolves anything needing a
// database.
package build

import (
	"encoding/json"
	"errors"
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
	names := config.NamedPolicyRefNames(role.NamedPolicies)
	if names == nil {
		names = []string{}
	}
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
		NamedPolicyNames:        names,
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

func NamedIAMPolicies(policies []config.NamedIAMPolicy, appConfigID string) []app.AppNamedIAMPolicyConfig {
	out := make([]app.AppNamedIAMPolicyConfig, 0, len(policies))
	for _, policy := range policies {
		obj := app.AppNamedIAMPolicyConfig{
			AppConfigID:             appConfigID,
			Name:                    policy.Name,
			PolicyName:              policy.Name,
			Description:             policy.Description,
			Contents:                dbgenerics.ToJSON(policy.Contents),
			CloudFormationStackName: app.NamedIAMPolicyCloudFormationStackName(policy.Name),
		}
		out = append(out, obj)
	}
	return out
}

const (
	awsCloudFormationStackType  = "aws-cloudformation"
	maxManagedPolicyAttachments = 10
)

func ValidateNamedIAMPolicies(in PermissionsInput, roles []app.AppAWSIAMRoleConfig) error {
	named := in.Permissions.NamedPolicies
	if len(named) == 0 {
		for _, role := range roles {
			if len(role.NamedPolicyNames) > 0 {
				return fmt.Errorf("role %q references named policies %v but none are defined", role.Name, role.NamedPolicyNames)
			}
		}
		return nil
	}

	if in.StackType != awsCloudFormationStackType {
		return fmt.Errorf("named IAM policies require stack type %q, got %q", awsCloudFormationStackType, in.StackType)
	}

	ids := make(map[string]struct{}, len(named))
	for _, policy := range named {
		id := policy.Name
		if id == "" {
			return errors.New("named policy is missing a name")
		}
		if _, ok := ids[id]; ok {
			return fmt.Errorf("named policy %q is duplicated", id)
		}
		ids[id] = struct{}{}
		if err := validateIAMPolicyDocument(fmt.Sprintf("named policy %q", id), dbgenerics.ToJSON(policy.Contents), true); err != nil {
			return err
		}
	}

	for _, role := range roles {
		platform := strings.ToLower(role.CloudPlatform)
		if platform != "" && platform != "aws" && len(role.NamedPolicyNames) > 0 {
			return fmt.Errorf("role %q: named IAM policies are AWS-only (cloud_platform %q)", role.Name, role.CloudPlatform)
		}

		managedCount := 0
		for _, policy := range role.Policies {
			if policy.ManagedPolicyName != "" {
				managedCount++
			}
		}
		for _, ref := range role.NamedPolicyNames {
			if _, ok := ids[ref]; !ok {
				return fmt.Errorf("role %q references unknown named policy %q", role.Name, ref)
			}
			managedCount++
		}
		if managedCount > maxManagedPolicyAttachments {
			return fmt.Errorf("role %q attaches %d managed policies (AWS managed plus named); the default IAM quota is %d", role.Name, managedCount, maxManagedPolicyAttachments)
		}
	}

	return nil
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
			if err := validateIAMPolicyDocument(fmt.Sprintf("role %q policy %q", role.Name, policy.Name), policy.Contents, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateIAMPolicyDocument(label string, contents []byte, requireContents bool) error {
	if len(contents) == 0 {
		if requireContents {
			return fmt.Errorf("%s: contents are required", label)
		}
		return nil
	}
	var doc struct {
		Statement []map[string]json.RawMessage `json:"Statement"`
	}
	if err := json.Unmarshal(contents, &doc); err != nil {
		return fmt.Errorf("%s: contents must be a JSON IAM policy document: %w", label, err)
	}
	// An empty Statement list is allowed: Azure configs use it as a
	// placeholder because the real grants come from RBAC.
	for i, stmt := range doc.Statement {
		if _, ok := stmt["Effect"]; !ok {
			return fmt.Errorf("%s: Statement[%d] missing required Effect field", label, i)
		}
	}
	return nil
}
