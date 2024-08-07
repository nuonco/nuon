package schema

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/stoewer/go-strcase"

	"github.com/powertoolsdev/mono/pkg/config"
)

func reflector() (*jsonschema.Reflector, error) {
	r := new(jsonschema.Reflector)

	r.FieldNameTag = "mapstructure"
	r.RequiredFromJSONSchemaTags = true
	r.KeyNamer = strcase.SnakeCase

	return r, nil
}

// This is used when the entire config file is in a single file, and generally should only be used _after_ parsing.
func AppSchemaFlat() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppConfig{}), nil
}

// This is used when the schema is using sources
func AppSchemaSources() (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	sourceSchema := r.Reflect(config.AppSourceConfig{})
	sourcesSchema := r.Reflect(config.AppSourcesConfig{})
	schma := r.Reflect(config.AppConfig{})
	schma.Definitions["AppSourceConfig"] = sourceSchema
	schma.Definitions["AppSourcesConfig"] = sourcesSchema

	err = setSchemaWithAnyOfSource("runner", schma, "AppRunnerConfig")
	if err != nil {
		return nil, err
	}

	err = setSchemaWithAnyOfSource("sandbox", schma, "AppSandboxConfig")
	if err != nil {
		return nil, err
	}

	err = setSchemaWithAnyOfSources("inputs", schma, "AppInputConfig")
	if err != nil {
		return nil, err
	}

	err = setSchemaWithAnyOfSource("installer", schma, "InstallerConfig")
	if err != nil {
		return nil, err
	}

	err = setItemsSchemaWithAnyOfSources("components", schma, "Component")
	if err != nil {
		return nil, err
	}

	return schma, nil
}

func setSchemaWithAnyOfSource(propertyName string, schma *jsonschema.Schema, defName string) error {
	schemaProperty, found := schma.Definitions["AppConfig"].Properties.Get(propertyName)
	if !found {
		return fmt.Errorf("unable to find %s in schema", propertyName)
	}
	schemaProperty.Ref = ""
	schemaProperty.Type = "object"
	schemaProperty.AnyOf = []*jsonschema.Schema{
		{
			Ref:         fmt.Sprintf("#/$defs/%s", defName),
			Description: fmt.Sprintf("%s configuration object", defName),
		},
		{
			Ref:         "#/$defs/AppSourceConfig",
			Description: "Source configuration object",
		},
	}

	return nil
}

func setSchemaWithAnyOfSources(propertyName string, schma *jsonschema.Schema, defName string) error {
	schemaProperty, found := schma.Definitions["AppConfig"].Properties.Get(propertyName)
	if !found {
		return fmt.Errorf("unable to find %s in schema", propertyName)
	}
	schemaProperty.Ref = ""
	schemaProperty.Type = "object"
	schemaProperty.AnyOf = []*jsonschema.Schema{
		{
			Ref:         fmt.Sprintf("#/$defs/%s", defName),
			Description: fmt.Sprintf("%s configuration object", defName),
		},
		{
			Ref:         "#/$defs/AppSourcesConfig",
			Description: "Sources configuration object",
		},
	}

	return nil
}

func setItemsSchemaWithAnyOfSources(propertyName string, schma *jsonschema.Schema, defName string) error {
	schemaProperty, found := schma.Definitions["AppConfig"].Properties.Get(propertyName)
	if !found {
		return fmt.Errorf("unable to find %s in schema", propertyName)
	}
	schemaProperty.Items.Ref = ""
	schemaProperty.Items.AnyOf = []*jsonschema.Schema{
		{
			Ref:         fmt.Sprintf("#/$defs/%s", defName),
			Description: fmt.Sprintf("%s configuration object", defName),
		},
		{
			Ref:         "#/$defs/AppSourceConfig",
			Description: "Source configuration object",
		},
	}

	return nil
}
