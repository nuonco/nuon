package mappers

import (
	"slices"
	"strings"

	"github.com/invopop/jsonschema"
)

func BuildPropertyMap(schema *jsonschema.Schema) (map[string]map[string]*jsonschema.Schema, map[string]map[string]bool) {
	hierarchicalMap := make(map[string]map[string]*jsonschema.Schema)
	requiredMap := make(map[string]map[string]bool)

	if schema == nil {
		return hierarchicalMap, requiredMap
	}

	defsLookup := schema.Definitions

	rootSchema := schema
	if schema.Ref != "" {
		resolved := resolveRef(schema.Ref, defsLookup)
		if resolved != nil {
			rootSchema = resolved
		}
	}

	buildPropertyMapRecursive(rootSchema, "", hierarchicalMap, requiredMap, defsLookup)
	for _, s := range schema.AllOf {
		defsLookup := s.Definitions
		sh := s
		if s.Ref != "" {
			resolved := resolveRef(s.Ref, defsLookup)
			if resolved != nil {
				sh = resolved
			}
		}
		buildPropertyMapRecursive(sh, "", hierarchicalMap, requiredMap, defsLookup)
	}
	return hierarchicalMap, requiredMap
}

func buildPropertyMapRecursive(schema *jsonschema.Schema, currentPath string, hierarchicalMap map[string]map[string]*jsonschema.Schema, requiredMap map[string]map[string]bool, defsLookup map[string]*jsonschema.Schema) {
	ignoredOneOffs := []string{
		"component_type",
		"docker_build",
		"external_image",
		"helm_chart",
		"job",
		"kubernetes_manifest",
		"terraform_module",
	}
	if schema == nil {
		return
	}

	if schema.Properties == nil {
		return
	}

	if hierarchicalMap[currentPath] == nil {
		hierarchicalMap[currentPath] = make(map[string]*jsonschema.Schema)
	}

	if len(schema.Required) > 0 {
		if requiredMap[currentPath] == nil {
			requiredMap[currentPath] = make(map[string]bool)
		}
		for _, req := range schema.Required {
			requiredMap[currentPath][req] = true
		}
	}

	pair := schema.Properties.Oldest()
	for pair != nil {
		key := pair.Key
		prop := pair.Value

		if v, ok := prop.Extras["oneof_required"]; ok {
			if slices.Contains(ignoredOneOffs, v.(string)) {
				return
			}
		}

		hierarchicalMap[currentPath][key] = prop

		if prop.Ref != "" {
			refDef := resolveRef(prop.Ref, defsLookup)
			if refDef != nil {
				nestedPath := key
				if currentPath != "" {
					nestedPath = currentPath + "." + key
				}
				buildPropertyMapRecursive(refDef, nestedPath, hierarchicalMap, requiredMap, defsLookup)
			}
		} else if prop.Type == "array" && prop.Items != nil {
			if prop.Items.Ref != "" {
				refDef := resolveRef(prop.Items.Ref, defsLookup)
				if refDef != nil {
					nestedPath := key
					if currentPath != "" {
						nestedPath = currentPath + "." + key
					}
					buildPropertyMapRecursive(refDef, nestedPath, hierarchicalMap, requiredMap, defsLookup)
				}
			} else if prop.Items.Properties != nil {
				nestedPath := key
				if currentPath != "" {
					nestedPath = currentPath + "." + key
				}
				buildPropertyMapRecursive(prop.Items, nestedPath, hierarchicalMap, requiredMap, defsLookup)
			}
		} else if prop.Properties != nil {
			nestedPath := key
			if currentPath != "" {
				nestedPath = currentPath + "." + key
			}
			buildPropertyMapRecursive(prop, nestedPath, hierarchicalMap, requiredMap, defsLookup)
		}

		pair = pair.Next()
	}
}

func resolveRef(ref string, defsLookup map[string]*jsonschema.Schema) *jsonschema.Schema {
	ref = strings.TrimPrefix(ref, "#/definitions/")
	ref = strings.TrimPrefix(ref, "#/$defs/")

	return defsLookup[ref]
}
