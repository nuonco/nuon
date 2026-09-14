package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/generics"
)

type PermissionsRoleType string

const (
	PermissionsRoleTypeProvision   PermissionsRoleType = "provision"
	PermissionsRoleTypeDeprovision PermissionsRoleType = "deprovision"
	PermissionsRoleTypeMaintenance PermissionsRoleType = "maintenance"
	PermissionsRoleTypeCustom      PermissionsRoleType = "custom"
)

var AllPermissionsRoleTypes []PermissionsRoleType = []PermissionsRoleType{
	PermissionsRoleTypeMaintenance,
	PermissionsRoleTypeProvision,
	PermissionsRoleTypeDeprovision,
	PermissionsRoleTypeCustom,
}

type PermissionsConfig struct {
	ProvisionRole   *AppAWSIAMRole   `mapstructure:"provision_role,omitempty" toml:"provision_role,omitempty"`
	DeprovisionRole *AppAWSIAMRole   `mapstructure:"deprovision_role,omitempty" toml:"deprovision_role,omitempty"`
	MaintenanceRole *AppAWSIAMRole   `mapstructure:"maintenance_role,omitempty" toml:"maintenance_role,omitempty"`
	CustomRoles     []*AppAWSIAMRole `mapstructure:"custom_roles,omitempty" toml:"custom_roles,omitempty"`

	Roles []*AppAWSIAMRole `mapstructure:"roles,omitempty" toml:"roles,omitempty"`

	NamedPolicies []NamedIAMPolicy `mapstructure:"named_policies,omitempty" toml:"named_policies,omitempty"`
}

func (a PermissionsConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("provision_role").Short("provisioning IAM role").
		Long("IAM role used during initial provisioning of the install with permissions to set up resources").
		Field("deprovision_role").Short("deprovisioning IAM role").
		Long("IAM role used for tearing down the install and cleaning up resources").
		Field("maintenance_role").Short("maintenance IAM role").
		Long("IAM role used for day-to-day maintenance, updates, and operational tasks").
		Field("custom_roles").Short("custom IAM roles").
		Long("Additional IAM roles for specialized operations beyond the standard provision/maintenance/deprovision lifecycle. Each role must have type set to 'custom'").
		Field("roles").Short("list of permission roles").
		Long("Array of role definitions in directory-based permission structure. Each role must have a type field (provision, maintenance, deprovision, or custom)").
		Field("named_policies").Short("named IAM policies").
		Long("Customer-managed IAM policies created independently of any role's Enable parameter. Attach them from a role with named_policies. Distinct from the root policies/ directory, which is Kyverno/OPA")
}

func (a *PermissionsConfig) parse() error {
	for _, role := range a.Roles {
		if role.Type == "" {
			return ErrConfig{
				Description: "role must have a type field when using directory structure",
				Err:         errors.New("role must have a type field when using directory"),
			}
		}

		if !generics.SliceContains(PermissionsRoleType(role.Type), AllPermissionsRoleTypes) {
			return ErrConfig{
				Description: fmt.Sprintf("role type must be one of (%s)", strings.Join(generics.ToStringSlice(AllPermissionsRoleTypes), ",")),
				Err:         errors.New("role has invalid type"),
			}
		}

		switch PermissionsRoleType(role.Type) {
		case PermissionsRoleTypeProvision:
			a.ProvisionRole = role
		case PermissionsRoleTypeDeprovision:
			a.DeprovisionRole = role
		case PermissionsRoleTypeMaintenance:
			a.MaintenanceRole = role
		case PermissionsRoleTypeCustom:
			a.CustomRoles = append(a.CustomRoles, role)
		}
	}

	return nil
}

func (a *PermissionsConfig) Validate() error {
	if a == nil {
		return errors.New("permissions config is required")
	}

	// validate duplicate role with same name
	roleNames := make(map[string]bool)
	for _, role := range a.Roles {
		if _, ok := roleNames[role.Name]; ok {
			return fmt.Errorf("role name %s is duplicated", role.Name)
		}

		roleNames[role.Name] = true
	}

	if a.ProvisionRole == nil {
		return errors.New("missing permission with type `provision`")
	}
	if a.MaintenanceRole == nil {
		return errors.New("missing permission with type `maintenance`")
	}
	if a.DeprovisionRole == nil {
		return errors.New("missing permission with type `deprovision`")
	}

	return validateNamedPolicyAttachments(a.NamedPolicies, a.allRoles())
}

func (a *PermissionsConfig) allRoles() []*AppAWSIAMRole {
	roles := make([]*AppAWSIAMRole, 0, 3+len(a.CustomRoles)+len(a.Roles))
	roles = append(roles, a.ProvisionRole, a.MaintenanceRole, a.DeprovisionRole)
	roles = append(roles, a.CustomRoles...)
	if len(a.Roles) > 0 {
		// Directory parse copies into the typed fields; Roles still holds the
		// same pointers. Deduplicate by identity so we don't double-check.
		seen := make(map[*AppAWSIAMRole]struct{}, len(roles))
		for _, role := range roles {
			if role != nil {
				seen[role] = struct{}{}
			}
		}
		for _, role := range a.Roles {
			if role == nil {
				continue
			}
			if _, ok := seen[role]; ok {
				continue
			}
			roles = append(roles, role)
		}
	}
	return roles
}

func validateNamedPolicyAttachments(named []NamedIAMPolicy, roles []*AppAWSIAMRole) error {
	if len(named) == 0 {
		for _, role := range roles {
			if role == nil {
				continue
			}
			if len(role.NamedPolicies) > 0 {
				return fmt.Errorf("role %q references named policies %v but none are defined", role.Name, NamedPolicyRefNames(role.NamedPolicies))
			}
		}
		return nil
	}

	names := make(map[string]struct{}, len(named))
	for _, policy := range named {
		if policy.Name == "" {
			return errors.New("named policy is missing a name")
		}
		if _, ok := names[policy.Name]; ok {
			return fmt.Errorf("named policy %s is duplicated", policy.Name)
		}
		names[policy.Name] = struct{}{}
	}

	for _, role := range roles {
		if role == nil {
			continue
		}
		for _, ref := range role.NamedPolicies {
			if _, ok := names[ref.Name]; !ok {
				return fmt.Errorf("role %q references unknown named policy %q", role.Name, ref.Name)
			}
		}
	}

	return nil
}
