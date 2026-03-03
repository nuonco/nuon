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
		Long("Array of role definitions in directory-based permission structure. Each role must have a type field (provision, maintenance, deprovision, or custom)")
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

	if a.ProvisionRole == nil || a.DeprovisionRole == nil || a.MaintenanceRole == nil {
		return errors.New("provision, deprovision and maintenance permissions are required")
	}
	return nil
}
