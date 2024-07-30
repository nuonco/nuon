package schema

import (
	"github.com/invopop/jsonschema"
	"github.com/stoewer/go-strcase"

	"github.com/powertoolsdev/mono/pkg/config"
)

const (
	defaultPackage = "github.com/powertoolsdev/mono/pkg/config"
)

func reflector() (*jsonschema.Reflector, error) {
	r := new(jsonschema.Reflector)

	r.FieldNameTag = "mapstructure"
	r.RequiredFromJSONSchemaTags = true
	r.KeyNamer = strcase.SnakeCase

	return r, nil
}

func AppSchema() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppConfig{}), nil
}

func AppSchemaSources() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	schma := r.Reflect(config.AppConfig{})

	runner, ok := schma.Definitions["AppRunnerConfig"]
	if ok {
		runner.Required = []string{"source"}
	}

	sandbox, ok := schma.Definitions["AppSandboxConfig"]
	if ok {
		sandbox.Required = []string{"source"}
	}

	inputs, ok := schma.Definitions["AppInputConfig"]
	if ok {
		inputs.Required = []string{}
	}

	installer, ok := schma.Definitions["InstallerConfig"]
	if ok {
		installer.Required = []string{"source"}
	}

	components, ok := schma.Definitions["Components"]
	if ok {
		components.Required = []string{"source"}
	}

	return schma, nil
}
