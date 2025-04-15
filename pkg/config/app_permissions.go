package config

import (
	"context"

	"github.com/invopop/jsonschema"
)

type PermissionsConfig struct {
	ProvisionRole   *AppAWSIAMRole `mapstructure:"provision_role" jsonschema:"required"`
	DeprovisionRole *AppAWSIAMRole `mapstructure:"deprovision_role" jsonschema:"required"`
	MaintenanceRole *AppAWSIAMRole `mapstructure:"maintenance_role" jsonschema:"required"`
}

func (a PermissionsConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "provision_role", "Role used during initial provisioning of the install.")
	addDescription(schema, "maintenance_role", "Role used for day to day maintenance and updates.")
	addDescription(schema, "deprovision_role", "Role used for tearing down the install.")
}

func (a *PermissionsConfig) parse(context.Context) error {
	return nil
}
