package config

import (
	"reflect"
)

func DecodeSource(fromType reflect.Type, toType reflect.Type, from interface{}) (interface{}, error) {
	return from, nil
	if fromType != reflect.TypeOf(map[string]interface{}{}) {
		return from, nil
	}
	if toType != reflect.TypeOf(Component{}) {
		return from, nil
	}

	return from, nil
}
