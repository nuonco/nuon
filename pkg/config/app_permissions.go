package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type PermissionsRoleType string

const (
	PermissionsRoleTypeProvision   PermissionsRoleType = "provision"
	PermissionsRoleTypeDeprovision PermissionsRoleType = "deprovision"
	PermissionsRoleTypeMaintenance PermissionsRoleType = "maintenance"
)

var AllPermissionsRoleTypes []PermissionsRoleType = []PermissionsRoleType{
	PermissionsRoleTypeMaintenance,
	PermissionsRoleTypeProvision,
	PermissionsRoleTypeDeprovision,
}

type PermissionsConfig struct {
	ProvisionRole   *AppAWSIAMRole `mapstructure:"provision_role,omitempty"`
	DeprovisionRole *AppAWSIAMRole `mapstructure:"deprovision_role,omitempty"`
	MaintenanceRole *AppAWSIAMRole `mapstructure:"maintenance_role,omitempty"`

	Roles []*AppAWSIAMRole `mapstructure:"roles,omitempty"`
}

func (a PermissionsConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("provision_role").Short("provisioning IAM role").
		Long("IAM role used during initial provisioning of the install with permissions to set up resources").
		Field("deprovision_role").Short("deprovisioning IAM role").
		Long("IAM role used for tearing down the install and cleaning up resources").
		Field("maintenance_role").Short("maintenance IAM role").
		Long("IAM role used for day-to-day maintenance, updates, and operational tasks").
		Field("roles").Short("list of permission roles").
		Long("Array of role definitions in directory-based permission structure. Each role must have a type field (provision, maintenance, or deprovision)")
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
		}
	}

	return nil
}
