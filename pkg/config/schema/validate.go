package schema

import (
	"context"

	"github.com/xeipuuv/gojsonschema"

	"github.com/powertoolsdev/mono/pkg/config"
)

func Validate(ctx context.Context, obj *config.AppConfig) ([]gojsonschema.ResultError, error) {
	jsonBytes, err := config.ToJSON(obj)
	if err != nil {
		return nil, err
	}

	schma, err := AppConfigSchema()
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
