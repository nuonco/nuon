package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

type AppInput struct {
	Name        string `mapstructure:"name"`
	DisplayName string `mapstructure:"display_name"`
	Description string `mapstructure:"description"`
	Group       string `mapstructure:"group"`
	Default     string `mapstructure:"default,omitempty"`
	Required    bool   `mapstructure:"required,omitempty"`
	Sensitive   bool   `mapstructure:"sensitive"`
}

func (a AppInput) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "Input name, which is used to reference the input via variable templating.")
	addDescription(schema, "display_name", "Human readable name of the input which is rendered in the installer.")
	addDescription(schema, "description", "Input description rendered in the installer.")
	addDescription(schema, "group", "The name of the group this field belongs too.")

	addDescription(schema, "default", "The default value for the input.")
	addDescription(schema, "required", "Denote whether this is a required customer input.")
	addDescription(schema, "sensitive", "Denote whether this is a sensitive input, which will prevent the value from being displayed after the install is created.")
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

func (a *AppInputConfig) parse(ctx ConfigContext) error {
	if ctx != ConfigContextSource {
		return nil
	}

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

	return nil
}
