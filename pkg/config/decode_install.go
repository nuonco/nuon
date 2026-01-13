package config

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// DecodeInstallInputs decodes inputs for an install.
func DecodeInstallInputs(fromType reflect.Type, toType reflect.Type, from interface{}) (interface{}, error) {
	if fromType != reflect.TypeOf([]interface{}{}) {
		return from, nil
	}
	if toType != reflect.TypeOf([]InputGroup{}) {
		return from, nil
	}

	var list []map[string]string
	err := mapstructure.Decode(from, &list)
	if err != nil {
		var group map[string]string
		err = mapstructure.Decode(from, &group)
		if err != nil {
			return from, fmt.Errorf("unable to convert install inputs: %w", err)
		}
		return []InputGroup{{Inputs: group}}, nil
	}

	result := make([]InputGroup, 0, len(list))
	for _, item := range list {
		result = append(result, InputGroup{Inputs: item})
	}

	return result, nil
}
