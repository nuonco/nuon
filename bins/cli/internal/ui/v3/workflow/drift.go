package workflow

import (
	"fmt"
	"reflect"
	"strings"
)

func interfaceToMap(data interface{}) (map[string]any, error) {
	// Check if data is nil
	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}

	// Try to assert directly to map[string]any
	if m, ok := data.(map[string]any); ok {
		return m, nil
	}

	// Try to assert to map[string]interface{}
	if m, ok := data.(map[string]interface{}); ok {
		return m, nil
	}

	// Use reflection to handle structs and other types
	v := reflect.ValueOf(data)

	// Dereference pointer if needed
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle map types
	if v.Kind() == reflect.Map {
		result := make(map[string]any)
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			// Ensure key is a string
			if key.Kind() != reflect.String {
				return nil, fmt.Errorf("map key is not a string: %v", key.Kind())
			}
			result[key.String()] = iter.Value().Interface()
		}
		return result, nil
	}

	// Handle struct types
	if v.Kind() == reflect.Struct {
		result := make(map[string]any)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			// Skip unexported fields
			if !field.IsExported() {
				continue
			}
			// Use json tag if available, otherwise use field name
			fieldName := field.Name
			if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
				// Handle json tag options like "field_name,omitempty"
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
