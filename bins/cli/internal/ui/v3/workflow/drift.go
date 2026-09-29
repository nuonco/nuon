package workflow

import (
	"fmt"
	"reflect"
	"strings"
)

func interfaceToMap(data interface{}) (map[string]any, error) {
	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}

	if m, ok := data.(map[string]any); ok {
		return m, nil
	}

	if m, ok := data.(map[string]interface{}); ok {
		return m, nil
	}

	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Map {
		result := make(map[string]any)
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				return nil, fmt.Errorf("map key is not a string: %v", key.Kind())
			}
			result[key.String()] = iter.Value().Interface()
		}
		return result, nil
	}

	if v.Kind() == reflect.Struct {
		result := make(map[string]any)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			fieldName := field.Name
			if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
				if idx := strings.Index(tag, ","); idx != -1 {
					fieldName = tag[:idx]
				} else {
					fieldName = tag
				}
			}
			result[fieldName] = v.Field(i).Interface()
		}
		return result, nil
	}

	return nil, fmt.Errorf("cannot convert type %T to map[string]any", data)
}
