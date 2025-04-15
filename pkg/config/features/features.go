package features

import (
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type FieldFeatures struct {
	Template bool `mapstructure:"template"`
	Get      bool `mapstructure:"get"`
}

func ParseFieldFeatures(field reflect.StructField) (*FieldFeatures, error) {
	var feats FieldFeatures

	// Assuming features are stored in a "features" tag
	features := field.Tag.Get("features")
	if features == "" {
		return &feats, nil
	}

	// Split features by comma and check for "gettable"
	optMap := generics.SliceToMapDefault[string, bool](strings.Split(features, ","), true)

	if err := mapstructure.Decode(optMap, &feats); err != nil {
		return nil, errors.Wrap(err, "unable to decode map")
	}

	return &feats, nil
}

func HasGetFeature(field reflect.StructField) (bool, error) {
	feats, err := ParseFieldFeatures(field)
	if err != nil {
		return false, errors.Wrap(err, "unable to parse features")
	}

	return feats.Get, nil
}

func HasTemplateFeature(field reflect.StructField) (bool, error) {
	feats, err := ParseFieldFeatures(field)
	if err != nil {
		return false, errors.Wrap(err, "unable to parse features")
	}

	return feats.Template, nil
}
