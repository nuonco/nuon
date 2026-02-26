package preflight

import (
	"fmt"
	"reflect"
	"strings"
)

// fieldInfo holds metadata about a config field tagged with preflight.
type fieldInfo struct {
	Name       string // Go struct field name
	ConfigName string // config tag value (env var name)
	Required   bool   // whether validate:"required" is present
	IsZero     bool   // whether the field has its zero value
}

// ValidateConfigForCheck inspects cfg for fields tagged preflight:"<check>"
// and returns an error listing any required fields that are missing (zero-valued).
func ValidateConfigForCheck(cfg any, check string) ([]fieldInfo, error) {
	fields := fieldsForCheck(cfg, check)
	if len(fields) == 0 {
		return nil, fmt.Errorf("no config fields tagged for check %q", check)
	}

	var missing []string
	for _, f := range fields {
		if f.Required && f.IsZero {
			missing = append(missing, f.ConfigName)
		}
	}

	if len(missing) > 0 {
		return fields, fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}

	return fields, nil
}

// ListChecks returns all distinct preflight check names found in struct tags.
func ListChecks(cfg any) []string {
	seen := map[string]bool{}
	var checks []string

	walkFields(cfg, func(field reflect.StructField, _ reflect.Value) {
		tag := field.Tag.Get("preflight")
		if tag == "" {
			return
		}
		for _, name := range strings.Split(tag, ",") {
			name = strings.TrimSpace(name)
			if name != "" && !seen[name] {
				seen[name] = true
				checks = append(checks, name)
			}
		}
	})

	return checks
}

// fieldsForCheck returns metadata for all fields belonging to a given check.
func fieldsForCheck(cfg any, check string) []fieldInfo {
	var fields []fieldInfo

	walkFields(cfg, func(field reflect.StructField, val reflect.Value) {
		tag := field.Tag.Get("preflight")
		if tag == "" {
			return
		}

		for _, name := range strings.Split(tag, ",") {
			if strings.TrimSpace(name) == check {
				configTag := field.Tag.Get("config")
				if configTag == "" || configTag == ",squash" {
					configTag = field.Name
				}

				validateTag := field.Tag.Get("validate")
				required := strings.Contains(validateTag, "required")

				fields = append(fields, fieldInfo{
					Name:       field.Name,
					ConfigName: configTag,
					Required:   required,
					IsZero:     val.IsZero(),
				})
				break
			}
		}
	})

	return fields
}

// walkFields iterates over all fields in a struct (including embedded/squashed structs).
func walkFields(cfg any, fn func(reflect.StructField, reflect.Value)) {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		val := v.Field(i)

		// Recurse into embedded structs (config:",squash" pattern).
		if field.Anonymous && val.Kind() == reflect.Struct {
			walkFields(val.Interface(), fn)
			continue
		}

		fn(field, val)
	}
}
