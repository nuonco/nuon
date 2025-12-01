package config

import (
	"reflect"
)

// DecodeSource is a global decoder supporting all sources
func DecodeSource(fromType reflect.Type, toType reflect.Type, from interface{}) (interface{}, error) {
	return from, nil
}
