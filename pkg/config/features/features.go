package features

import (
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
)

type FieldFeatures struct {
	Template bool `mapstructure:"template" toml:"template"`
	Get      bool `mapstructure:"get" toml:"get"`
}

func ParseFieldFeatures(field reflect.StructField) (*FieldFeatures, error) {
	var feats FieldFeatures

	features := field.Tag.Get("features")
	if features == "" {
		return &feats, nil
	}

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
