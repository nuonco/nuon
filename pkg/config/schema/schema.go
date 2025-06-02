package schema

import (
	"github.com/invopop/jsonschema"
	"github.com/stoewer/go-strcase"
)

func reflector() (*jsonschema.Reflector, error) {
	r := new(jsonschema.Reflector)

	r.FieldNameTag = "mapstructure"
	r.RequiredFromJSONSchemaTags = true
	r.KeyNamer = strcase.SnakeCase

	return r, nil
}
