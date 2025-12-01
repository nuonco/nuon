package mappers

import (
	"strings"

	"github.com/invopop/jsonschema"
)

// BuildPropertyMap recursively collects all properties from a JSON schema into a hierarchical map.
// The outer map key is the table path (empty string "" for root level).
// The inner map contains properties available at that level.
// Resolves $ref pointers to definitions to properly handle nested structures.
// This is useful for context-aware property lookups in LSP features like hover, completion, etc.
func BuildPropertyMap(schema *jsonschema.Schema) map[string]map[string]*jsonschema.Schema {
	hierarchicalMap := make(map[string]map[string]*jsonschema.Schema)

	if schema == nil {
		return hierarchicalMap
	}

	// Use schema.Definitions directly as the lookup map for $ref resolution
	// (Definitions is already a map[string]*Schema)
	defsLookup := schema.Definitions

	// If the schema root has a $ref, resolve it and start from there
	rootSchema := schema
	if schema.Ref != "" {
		resolved := resolveRef(schema.Ref, defsLookup)
		if resolved != nil {
			rootSchema = resolved
		}
	}

	buildPropertyMapRecursive(rootSchema, "", hierarchicalMap, defsLookup)
	return hierarchicalMap
}

// buildPropertyMapRecursive is the internal recursive function that populates the hierarchical property map
// currentPath represents the dotted path to the current level (e.g., "public_repo" or "values_file")
// defsLookup contains all schema definitions for $ref resolution
func buildPropertyMapRecursive(schema *jsonschema.Schema, currentPath string, hierarchicalMap map[string]map[string]*jsonschema.Schema, defsLookup map[string]*jsonschema.Schema) {
	if schema == nil {
		return
	}

	// Process regular properties
	if schema.Properties == nil {
		return
	}

	// Initialize the map for this path level if it doesn't exist
	if hierarchicalMap[currentPath] == nil {
		hierarchicalMap[currentPath] = make(map[string]*jsonschema.Schema)
	}

	pair := schema.Properties.Oldest()
	for pair != nil {
		key := pair.Key
		prop := pair.Value

		// Store property at current level
		hierarchicalMap[currentPath][key] = prop

		// Handle nested properties - either inline, via $ref, or in array items
		if prop.Ref != "" {
			// Resolve $ref and process the referenced definition
			refDef := resolveRef(prop.Ref, defsLookup)
			if refDef != nil {
				// Build the nested path
				nestedPath := key
				if currentPath != "" {
					nestedPath = currentPath + "." + key
				}
				buildPropertyMapRecursive(refDef, nestedPath, hierarchicalMap, defsLookup)
			}
		} else if prop.Type == "array" && prop.Items != nil {
			// Handle array types - check if items have $ref or properties
			if prop.Items.Ref != "" {
				refDef := resolveRef(prop.Items.Ref, defsLookup)
				if refDef != nil {
					nestedPath := key
					if currentPath != "" {
						nestedPath = currentPath + "." + key
					}
					buildPropertyMapRecursive(refDef, nestedPath, hierarchicalMap, defsLookup)
				}
			} else if prop.Items.Properties != nil {
				nestedPath := key
				if currentPath != "" {
					nestedPath = currentPath + "." + key
				}
				buildPropertyMapRecursive(prop.Items, nestedPath, hierarchicalMap, defsLookup)
			}
		} else if prop.Properties != nil {
			// Handle inline nested properties
			nestedPath := key
			if currentPath != "" {
				nestedPath = currentPath + "." + key
			}
			buildPropertyMapRecursive(prop, nestedPath, hierarchicalMap, defsLookup)
		}

		pair = pair.Next()
	}
}

// resolveRef resolves a JSON Schema $ref to its definition
// Supports formats like "#/definitions/TypeName" or "#/$defs/TypeName"
func resolveRef(ref string, defsLookup map[string]*jsonschema.Schema) *jsonschema.Schema {
	// Handle common $ref formats: "#/definitions/TypeName" or "#/$defs/TypeName"
	ref = strings.TrimPrefix(ref, "#/definitions/")
	ref = strings.TrimPrefix(ref, "#/$defs/")

	return defsLookup[ref]
}
