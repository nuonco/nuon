package schema

import (
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/xeipuuv/gojsonschema"
)

func Validate(obj *config.AppConfig) ([]gojsonschema.ResultError, error) {
	jsonBytes, err := config.ToJSON(obj)
	if err != nil {
		return nil, err
	}

	schma, err := AppSchemaFlat()
	if err != nil {
		return nil, err
	}

	schmaLoader := gojsonschema.NewGoLoader(schma)
	docLoader := gojsonschema.NewStringLoader(string(jsonBytes))

	res, err := gojsonschema.Validate(schmaLoader, docLoader)
	if err != nil {
		return nil, err
	}

	return res.Errors(), nil
}
