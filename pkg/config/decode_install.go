package config

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// DecodeInstallInputs decodes inputs for an install.
func DecodeInstallInputs(fromType reflect.Type, toType reflect.Type, from interface{}) (interface{}, error) {
	if fromType != reflect.TypeOf(map[string]interface{}{}) {
		return from, nil
	}
	if toType != reflect.TypeOf([]InputGroup{}) {
		return from, nil
	}

	var list []InputGroup
	err := mapstructure.Decode(from, &list)
	if err != nil {
		var group InputGroup
		err = mapstructure.Decode(from, &group)
		if err != nil {
			return from, fmt.Errorf("unable to convert install inputs: %w", err)
		}
		return []InputGroup{group}, nil
	}

	return list, nil
}
