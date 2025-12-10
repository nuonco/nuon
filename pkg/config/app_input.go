package config

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
	"github.com/powertoolsdev/mono/pkg/generics"
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
	NewSchemaBuilder(schema).
		Field("name").Short("input name").
		Long("Used to reference the input via variable templating (e.g., {{.nuon.inputs.input_name}})").
		Example("api_token").
		Example("database_url").
		Field("display_name").Short("display name of the input").Required().
		Long("Human-readable name shown in the installer UI to customers").
		Example("API Token").
		Example("Database URL").
		Field("description").Short("input description").Required().
		Long("Detailed explanation of what this input is for, rendered in the installer to guide users").
		Example("The API token for authenticating with the external service").
		Example("Connection string for the PostgreSQL database").
		Field("group").Short("input group name").Required().
		Long("Name of the input group this field belongs to. Must match a defined group in the inputs section").
		Example("database").
		Example("integrations").
		Field("default").Short("default value for the input").
		Long("Default value used if customer does not provide one. Type must match the input type").
		Example("production").
		Example("5432").
		Field("required").Short("whether input is required").
		Long("If true, customer must provide a value during installation. If false, can be skipped").
		Field("sensitive").Short("whether input is sensitive").
		Long("If true, the value will be masked/hidden in the UI and logs after the install is created. Use for passwords, tokens, and API keys").
		Field("type").Short("input type").
		Long("Data type for the input. Supported types: string, number, list, json, bool").
		Example("string").
		Example("number").
		Example("json").
		Example("bool").
		Field("internal").Short("whether input is internal-only").
		Long("If true, input is only settable via the admin panel and not shown to regular users").
		Field("user_configurable").Short("whether input is user configurable").
		Long("If true, input can be modified by end users after installation")
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
	NewSchemaBuilder(schema).
		Field("input").Short("list of inputs").
		Long("Array of input definitions that customers can configure during installation").
		Field("group").Short("list of input groups").
		Long("Array of input group definitions that organize related inputs in the installer UI")
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
	grps := make([]string, 0)
	for _, grp := range a.Groups {
		grps = append(grps, grp.Name)
	}

	for _, input := range a.Inputs {
		if input.Group != "" && !generics.SliceContains(input.Group, grps) {
			return ErrConfig{
				Description: fmt.Sprintf("input %s specified a group that does not exist %s", input.Name, input.Group),
				Err:         fmt.Errorf("group %s does not exist", input.Group),
			}
		}

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
