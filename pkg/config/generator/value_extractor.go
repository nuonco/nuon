package generator

import (
	"fmt"
	"reflect"
	"strings"
)

type InstanceValueExtractor struct {
	instance     any
	reflectValue reflect.Value
}

func NewInstanceValueExtractor(instance any) *InstanceValueExtractor {
	if instance == nil {
		return nil
	}
	return &InstanceValueExtractor{
		instance:     instance,
		reflectValue: reflect.ValueOf(instance),
	}
}

func (e *InstanceValueExtractor) GetFieldValue(path string) (any, bool) {
	if e != nil && !e.reflectValue.IsValid() {
		return nil, false
	}

	pathParts := strings.Split(path, ".")
	current := e.reflectValue

	for _, part := range pathParts {
		if current.Kind() == reflect.Pointer {
			if current.IsNil() {
				return nil, false
			}

			current = current.Elem()
		}

		if current.Kind() == reflect.Struct {
			structField, ok := e.findFieldByMapstructureTag(current.Type(), part)
			if !ok {
				return nil, false
			}
			current = current.FieldByIndex(structField.Index)
		} else {
			return nil, false
		}
	}

	if current.Kind() == reflect.Pointer {
		if current.IsNil() {
			return nil, false
		}

		current = current.Elem()
	}

	if current.IsZero() {
		return nil, false
	}

	return current.Interface(), true
}

func (e *InstanceValueExtractor) HasValue(propertyPath string) bool {
	_, exists := e.GetFieldValue(propertyPath)
	return exists
}

func (e *InstanceValueExtractor) findFieldByMapstructureTag(t reflect.Type, tagValue string) (reflect.StructField, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if vs, ok := field.Tag.Lookup("mapstructure"); ok {
			if strings.Split(vs, ",")[0] == tagValue {
				return field, true
			}
		}

		if strings.EqualFold(field.Name, tagValue) {
			return field, true
		}
	}

	return reflect.StructField{}, false
}

type ArrayElementType string

const (
	ArrayElementTypePremitive = "premitive"
	ArrayElementtypeObject    = "object"
)

func (e *InstanceValueExtractor) GetArrayValue(propertyPath string) (reflect.Value, string, bool) {
	field, ok := e.GetFieldValue(propertyPath)
	if !ok {
		return reflect.Value{}, "", false
	}

	fieldValue := reflect.ValueOf(field)
	if fieldValue.Kind() != reflect.Slice && fieldValue.Kind() != reflect.Array {
		return reflect.Value{}, "", false
	}

	if fieldValue.Len() == 0 {
		return fieldValue, "", true
	}

	item := fieldValue.Index(0)
	if item.Kind() == reflect.Pointer && !item.IsNil() {
		item = item.Elem()
	}

	if item.Kind() == reflect.Struct || item.Kind() == reflect.Map {
		return fieldValue, ArrayElementtypeObject, true
	}

	return fieldValue, ArrayElementTypePremitive, true
}

func (e *InstanceValueExtractor) GetMapValue(propertyPath string) (map[string]string, bool) {
	field, ok := e.GetFieldValue(propertyPath)
	if !ok {
		return nil, false
	}

	fieldValue := reflect.ValueOf(field)
	if fieldValue.Kind() != reflect.Map {
		return nil, false
	}

	if fieldValue.Len() == 0 {
		return nil, false
	}

	result := make(map[string]string)
	iter := fieldValue.MapRange()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		keyStr := fmt.Sprintf("%v", key.Interface())
		valueStr := fmt.Sprintf("%v", value.Interface())

		result[keyStr] = valueStr
	}

	return result, true
}
