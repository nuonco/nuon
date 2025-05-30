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

// StructHasher implements reflectwalk.StructWalker to collect hashable field data
type StructHasher struct {
	fieldData []string
	path      []string
}

// Struct is called for each struct encountered during the walk
func (s *StructHasher) Struct(v reflect.Value) error {
	return nil
}

// StructField is called for each field in a struct
func (s *StructHasher) StructField(field reflect.StructField, v reflect.Value) error {
	// Skip unexported fields
	if !field.IsExported() {
		return reflectwalk.SkipEntry
	}

	// Check nuonhash tag for exclusion
	tag := field.Tag.Get("nuonhash")
	if tag == "-" {
		return reflectwalk.SkipEntry
	}

	// Get the field name (use tag name if available, otherwise convert field name to snake_case)
	fieldName := toSnakeCase(field.Name)
	if tag != "" && !strings.Contains(tag, ",") {
		fieldName = tag
	} else if strings.Contains(tag, ",") {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			fieldName = parts[0]
		}
	}

	// Build the full path
	fullPath := strings.Join(append(s.path, fieldName), ".")

	// For primitive types, add to field data
	if s.isPrimitive(v) {
		fieldStr := fmt.Sprintf("%s:%v", fullPath, v.Interface())
		s.fieldData = append(s.fieldData, fieldStr)
		return reflectwalk.SkipEntry // Don't walk into primitive values
	}

	// For non-primitives, add the field name to path and continue walking
	s.path = append(s.path, fieldName)
	return nil
}

// Enter is called when entering a new level during walk
func (s *StructHasher) Enter(location reflectwalk.Location) error {
	return nil
}

// Exit is called when exiting a level during walk
func (s *StructHasher) Exit(location reflectwalk.Location) error {
	// Pop the last path element when exiting a struct field
	if location == reflectwalk.StructField && len(s.path) > 0 {
		s.path = s.path[:len(s.path)-1]
	}
	return nil
}

// Slice handles slice entries
func (s *StructHasher) Slice(v reflect.Value) error {
	return nil
}

// SliceElem handles individual slice elements
func (s *StructHasher) SliceElem(i int, v reflect.Value) error {
	if s.isPrimitive(v) {
		// For slices of primitives, include the index in the path
		currentPath := strings.Join(s.path, ".")
		fieldStr := fmt.Sprintf("%s[%d]:%v", currentPath, i, v.Interface())
		s.fieldData = append(s.fieldData, fieldStr)
		return nil // Don't use SkipEntry here
	}

	// For slices of structs, add index to path
	indexedPath := fmt.Sprintf("%s[%d]", strings.Join(s.path, "."), i)
	s.path = []string{indexedPath}
	return nil
}

// Array handles array entries
func (s *StructHasher) Array(v reflect.Value) error {
	return nil
}

// ArrayElem handles individual array elements
func (s *StructHasher) ArrayElem(i int, v reflect.Value) error {
	return s.SliceElem(i, v) // Same logic as slice elements
}

// Map handles map entries
func (s *StructHasher) Map(v reflect.Value) error {
	// Let reflectwalk handle the map iteration, but we'll sort the results later
	return nil
}

// MapElem handles individual map elements
func (s *StructHasher) MapElem(m, k, v reflect.Value) error {
	if s.isPrimitive(v) {
		currentPath := strings.Join(s.path, ".")
		fieldStr := fmt.Sprintf("%s[%v]:%v", currentPath, k.Interface(), v.Interface())
		s.fieldData = append(s.fieldData, fieldStr)
		return nil // Don't use SkipEntry here
	}

	// For maps with struct values, add key to path
	keyedPath := fmt.Sprintf("%s[%v]", strings.Join(s.path, "."), k.Interface())
	s.path = []string{keyedPath}
	return nil
}

// Primitive handles primitive value types
func (s *StructHasher) Primitive(v reflect.Value) error {
	// This shouldn't be called directly since we handle primitives in StructField
	return nil
}

// isPrimitive checks if a value is a primitive type that should be included directly
func (s *StructHasher) isPrimitive(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	case reflect.Ptr:
		// Handle nil pointers as primitives
		if v.IsNil() {
			return true
		}
		// For non-nil pointers, check the underlying type
		return s.isPrimitive(v.Elem())
	default:
		return false
	}
}

// toSnakeCase converts a string from PascalCase/camelCase to snake_case
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

// HashStruct creates a hash of a struct using reflectwalk, ignoring fields marked with `-` in the nuonhash tag
func HashStruct(v interface{}) (string, error) {
	hasher := &StructHasher{
		fieldData: make([]string, 0),
		path:      make([]string, 0),
	}

	// Walk the struct
	err := reflectwalk.Walk(v, hasher)
	if err != nil {
		return "", fmt.Errorf("error walking struct: %w", err)
	}

	// Sort field data for consistent hashing
	sort.Strings(hasher.fieldData)

	// Create hash
	hash := sha256.New()
	for _, data := range hasher.fieldData {
		hash.Write([]byte(data))
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
