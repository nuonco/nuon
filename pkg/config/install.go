package config

import "github.com/invopop/jsonschema"

// Install is a flattened configuration type that allows us to define installs for an app.
type Install struct {
	Name      string            `mapstructure:"name" jsonschema:"required"`
	AWSRegion string            `mapstructure:"aws_region,omitempty"`
	Inputs    map[string]string `mapstructure:"inputs,omitempty"`
}

func (a Install) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "name of the install")
	addDescription(schema, "aws_region", "AWS region to use for the install, if not set, the default region will be used")
	addDescription(schema, "inputs", "list of inputs")
}

func (i *Install) parse() error {
	if i == nil {
		return nil
	}

	return nil
}

func (i *Install) Validate() error {
	return nil
}
