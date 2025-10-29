package config

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

type AppInputSource string

const (
	AppInputSourceVendor   AppInputSource = "vendor"
	AppInputSourceCustomer AppInputSource = "customer"
)

type AppInput struct {
	Name             string `mapstructure:"name"`
	DisplayName      string `mapstructure:"display_name" jsonschema:"required"`
	Description      string `mapstructure:"description" jsonschema:"required"`
	Group            string `mapstructure:"group" jsonschema:"required"`
	Default          any    `mapstructure:"default,omitempty"`
	Required         bool   `mapstructure:"required,omitempty"`
	Sensitive        bool   `mapstructure:"sensitive"`
	Type             string `mapstructure:"type"`
	Internal         bool   `mapstructure:"internal"`
	UserConfigurable bool   `mapstructure:"user_configurable"`
}

func (a AppInput) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "Input name, which is used to reference the input via variable templating.")
	addDescription(schema, "display_name", "Human readable name of the input which is rendered in the installer.")
	addDescription(schema, "description", "Input description rendered in the installer.")
	addDescription(schema, "group", "The name of the group this field belongs too.")

	addDescription(schema, "default", "The default value for the input.")
	addDescription(schema, "required", "Denote whether this is a required customer input.")
	addDescription(schema, "sensitive", "Denote whether this is a sensitive input, which will prevent the value from being displayed after the install is created.")
	addDescription(schema, "type", "Type of input supported, can be a string, number, list, json or bool")
	addDescription(schema, "internal", "Internal inputs are only settable via the admin panel")
	addDescription(schema, "source", "Source of the input value. Can be 'user' (default, provided during install creation) or 'install_stack' (provided by CloudFormation/Bicep stack outputs via phone home).")
}

type AppInputGroup struct {
	Name        string `mapstructure:"name" jsonschema:"required"`
	Description string `mapstructure:"description" jsonschema:"required"`
	DisplayName string `mapstructure:"display_name,omitempty"`
}

func (a AppInputGroup) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "Group name, which must be referenced by each input.")
	addDescription(schema, "description", "Human readable description which is rendered in the installer.")
	addDescription(schema, "display_name", "Human readable name which is rendered in the installer.")
}

type AppInputConfig struct {
	Inputs []AppInput      `mapstructure:"input,omitempty" toml:"input"`
	Groups []AppInputGroup `mapstructure:"group,omitempty" toml:"group"`

	Source  string   `mapstructure:"source,omitempty"`
	Sources []string `mapstructure:"sources,omitempty"`
}

func (a AppInputConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "inputs", "list of inputs")
	addDescription(schema, "groups", "list of input groups")
}

func (a *AppInputConfig) parse() error {
	sources := make([]string, 0)
	if a.Source != "" {
		sources = append(sources, a.Source)
	}
	sources = append(sources, a.Sources...)

	for _, src := range sources {
		obj, err := source.LoadSource(src)
		if err != nil {
			return ErrConfig{
				Description: fmt.Sprintf("unable to load source %s", src),
				Err:         err,
			}
		}

		var inpCfg AppInputConfig
		if err := mapstructure.Decode(obj, &inpCfg); err != nil {
			return fmt.Errorf("unable to parse input source %s: %w", src, err)
		}

		a.Inputs = append(a.Inputs, inpCfg.Inputs...)
		a.Groups = append(a.Groups, inpCfg.Groups...)
	}

	err := a.ValidateInputs()
	if err != nil {
		return err
	}

	return nil
}

func (a *AppInputConfig) ValidateInputs() error {
	for _, input := range a.Inputs {
		if input.Type == "json" {
			if input.Default != nil {
				if _, ok := input.Default.(string); !ok {
					return ErrConfig{
						Description: fmt.Sprintf("input %s has a default value that is not a json string", input.Name),
						Err:         fmt.Errorf("input %s default value must be a json string", input.Name),
					}
				}
				if !json.Valid([]byte(input.Default.(string))) {
					return ErrConfig{
						Description: fmt.Sprintf("input %s has an invalid JSON string", input.Name),
						Err:         fmt.Errorf("input %s default value is not valid JSON string", input.Name),
					}
				}
			}
		}
	}

	return nil
}
