package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/mitchellh/reflectwalk"
)

type StructHasherOptions struct {
	EnableOmitEmpty bool
}

type StructHasher struct {
	fieldData []string
	path      []string

	options StructHasherOptions
}

func (s *StructHasher) Struct(v reflect.Value) error {
	return nil
}

func (s *StructHasher) StructField(field reflect.StructField, v reflect.Value) error {
	if !field.IsExported() {
		return reflectwalk.SkipEntry
	}

	tag := field.Tag.Get("nuonhash")
	fieldName := toSnakeCase(field.Name)
	omitEmpty := false

	if tag != "" {
		parts := strings.Split(tag, ",")

		for _, part := range parts {
			switch strings.TrimSpace(part) {
			case "-":
				return reflectwalk.SkipEntry
			case "omitempty":
				if s.options.EnableOmitEmpty {
					omitEmpty = true
					continue
				}
			}
		}
	}

	if omitEmpty && s.isEmpty(v) {
		return reflectwalk.SkipEntry
	}

	fullPath := strings.Join(append(s.path, fieldName), ".")

	if s.isPrimitive(v) {
		fieldStr := fmt.Sprintf("%s:%v", fullPath, s.formatFieldValue(v))
		s.fieldData = append(s.fieldData, fieldStr)
		return reflectwalk.SkipEntry
	}

	s.path = append(s.path, fieldName)
	return nil
}

func (s *StructHasher) isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		return v.IsZero()
	}
	return false
}

func (s *StructHasher) Enter(location reflectwalk.Location) error {
	return nil
}

func (s *StructHasher) Exit(location reflectwalk.Location) error {
	if location == reflectwalk.StructField && len(s.path) > 0 {
		s.path = s.path[:len(s.path)-1]
	}
	return nil
}

func (s *StructHasher) Slice(v reflect.Value) error {
	return nil
}

func (s *StructHasher) SliceElem(i int, v reflect.Value) error {
	if s.isPrimitive(v) {
		currentPath := strings.Join(s.path, ".")
		fieldStr := fmt.Sprintf("%s[%d]:%v", currentPath, i, s.formatFieldValue(v))
		s.fieldData = append(s.fieldData, fieldStr)
		return nil // why: Don't use SkipEntry here
	}

	indexedPath := fmt.Sprintf("%s[%d]", strings.Join(s.path, "."), i)
	s.path = []string{indexedPath}
	return nil
}

func (s *StructHasher) Array(v reflect.Value) error {
	return nil
}

func (s *StructHasher) ArrayElem(i int, v reflect.Value) error {
	return s.SliceElem(i, v)
}

func (s *StructHasher) Map(v reflect.Value) error {
	return nil
}

func (s *StructHasher) MapElem(m, k, v reflect.Value) error {
	if s.isPrimitive(v) {
		currentPath := strings.Join(s.path, ".")
		fieldStr := fmt.Sprintf("%s[%v]:%v", currentPath, k.Interface(), s.formatFieldValue(v))
		s.fieldData = append(s.fieldData, fieldStr)
		return nil // why: Don't use SkipEntry here
	}

	keyedPath := fmt.Sprintf("%s[%v]", strings.Join(s.path, "."), k.Interface())
	s.path = []string{keyedPath}
	return nil
}

func (s *StructHasher) Primitive(v reflect.Value) error {
	return nil
}

func (s *StructHasher) isPrimitive(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return s.isPrimitive(v.Elem())
	default:
		return false
	}
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func (s *StructHasher) formatFieldValue(v reflect.Value) any {
	if !v.IsValid() {
		return ""
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}

		elem := v.Elem()
		if !elem.IsValid() {
			return ""
		}

		return elem.Interface()
	}

	return v.Interface()
}

func HashStruct(v interface{}, options StructHasherOptions) (string, error) {
	hasher := &StructHasher{
		fieldData: make([]string, 0),
		path:      make([]string, 0),
		options:   options,
	}

	err := reflectwalk.Walk(v, hasher)
	if err != nil {
		return "", fmt.Errorf("error walking struct: %w", err)
	}

	sort.Strings(hasher.fieldData)

	hash := sha256.New()
	for _, data := range hasher.fieldData {
		hash.Write([]byte(data))
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
