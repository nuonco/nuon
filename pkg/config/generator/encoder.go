package generator

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/invopop/jsonschema"
)

// recursivelyEncode traverses a JSON schema and generates TOML configuration content.
// It handles nested objects, arrays, and primitive fields while respecting optional/required field rules.
//
// Parameters:
//   - schema: The JSON schema to process
//   - oneOfGroups: Map of oneOf group names to their required fields for conditional field inclusion, this is only passed from root not as part of
//     recursive calls
//   - output: String builder to write the generated TOML content to
//   - prefix: Current property path prefix for nested objects
//   - parentOptional: Whether the parent object is optional (unused in current implementation)
//   - writeComments: Whether to include property comments in the output
//   - skipNonRequired: Whether to skip non-required fields in the output
//   - extractor: Instance value extractor for retrieving actual values from struct instances
//
// Returns an error if the schema is invalid or if there are issues during encoding.
func (g *ConfigGen) recursivelyEncode(schema *jsonschema.Schema, oneOfGroups map[string]map[string]bool, output *strings.Builder, prefix string, parentOptional bool, writeComments bool, skipNonRequired bool, extractor *InstanceValueExtractor) error {
	if schema == nil || schema.Properties == nil {
		return fmt.Errorf("schema or properties is nil")
	}

	skipNonRequired = skipNonRequired || g.SkipNonRequired

	requiredFields := make(map[string]bool)
	for _, fieldName := range schema.Required {
		requiredFields[fieldName] = true
	}

	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		propertyName := pair.Key
		propertySchema := pair.Value

		isRequired := requiredFields[propertyName]

		fullPath := propertyName
		if prefix != "" {
			fullPath = prefix + "." + propertyName
		}

		hasInstanceValue := extractor != nil && extractor.HasValue(fullPath)

		isOptional := skipNonRequired && (!isRequired || parentOptional) && !hasInstanceValue

		if skipNonRequired && !isRequired && !hasInstanceValue {
			continue
		}

		if !g.EnableDeprecated && propertySchema.Deprecated {
			continue
		}

		oneOfRequiredGroup, ok := propertySchema.Extras[StructTagOneofRequired]
		// we check if the current property contains one of group extras properties
		// also if one of groups is present in recursive inputs
		if ok && oneOfGroups != nil && slices.Contains(StructTagOneOfRequiredGroups, oneOfRequiredGroup.(string)) {
			// if yes, we check if the property name is included in the
			oneOfGroupName := oneOfRequiredGroup.(string)
			// check if one of group name exists, if exits check if the property name exists in that group name, if yes, skip
			if oneOfGroup, ok := oneOfGroups[oneOfGroupName]; ok && oneOfGroup[propertyName] && !hasInstanceValue {
				continue
			}
		}

		if writeComments && g.EnableInfoComments {
			g.writePropertyComments(propertySchema, output)
		}

		// Handle different types
		switch propertySchema.Type {
		case "array":
			// array contents
			err := g.encodeTOMLArray(fullPath, propertySchema, output, isOptional, writeComments, extractor, fullPath)
			if err != nil {
				return err
			}
			output.WriteString("\n")
		case "object":
			// nested objects
			err := g.encodeTOMLObject(fullPath, propertySchema, output, isOptional, writeComments, extractor, fullPath)
			if err != nil {
				return err
			}
			output.WriteString("\n")
		default:
			g.writePrimitiveField(propertyName, propertySchema, output, isOptional, extractor, fullPath)
			if writeComments && g.EnableInfoComments {
				output.WriteString("\n")
			}

		}

		// output.WriteString("\n")
	}
	return nil
}

func (g *ConfigGen) writePropertyComments(schema *jsonschema.Schema, output *strings.Builder) {
	// short description
	if schema.Title != "" {
		output.WriteString("# ")
		output.WriteString(schema.Title)
		output.WriteString("\n")
	}

	// long description
	if schema.Description != "" {
		for line := range strings.SplitSeq(schema.Description, "\n") {
			output.WriteString("# ")
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// examples
	if len(schema.Examples) > 0 {
		output.WriteString("# Examples: ")
		for i, example := range schema.Examples {
			if i > 0 {
				output.WriteString(", ")
			}
			exampleStr := fmt.Sprintf("%v", example)
			// Handle multiline examples by writing them on new commented lines
			if strings.Contains(exampleStr, "\n") {
				output.WriteString("\n")
				for _, line := range strings.Split(exampleStr, "\n") {
					output.WriteString("# ")
					output.WriteString(line)
					output.WriteString("\n")
				}
			} else {
				output.WriteString(exampleStr)
			}
		}
		output.WriteString("\n")
	}
}

func (g *ConfigGen) writePrimitiveField(fieldName string, schema *jsonschema.Schema, output *strings.Builder, isOptional bool, extractor *InstanceValueExtractor, propertyPath string) {
	fieldLine := g.generateFieldLine(fieldName, schema, extractor, propertyPath)

	// If optional, comment it out
	if isOptional {
		output.WriteString("# ")
	}

	output.WriteString(fieldLine)
	// TODO(sk): remove this if comments are diabled
	output.WriteString("\n")
}

func (g *ConfigGen) encodeTOMLObject(tableName string, schema *jsonschema.Schema, output *strings.Builder, isOptional bool, writeComments bool, extractor *InstanceValueExtractor, propertyPath string) error {
	if schema.Properties == nil || schema.Properties.Len() == 0 {
		if extractor != nil {
			if mapValue, exists := extractor.GetMapValue(propertyPath); exists && len(mapValue) > 0 {
				commentPrefix := ""
				if isOptional {
					commentPrefix = "# "
				}

				fmt.Fprintf(output, "%s[%s]\n", commentPrefix, tableName)

				for key, value := range mapValue {
					fmt.Fprintf(output, "%s%s = \"%s\"\n", commentPrefix, key, value)
				}

				return nil
			}
		}

		if isOptional {
			output.WriteString("# ")
		}
		fmt.Fprintf(output, "# [%s] \n", tableName)
		output.WriteString("# key = \"value\" \n")

		return nil
	}

	commentPrefix := ""
	if isOptional {
		commentPrefix = "# "
	}

	fmt.Fprintf(output, "%s[%s]\n", commentPrefix, tableName)

	g.recursivelyEncode(schema, nil, output, tableName, isOptional, writeComments, false, extractor)
	return nil
}

func (g *ConfigGen) encodeTOMLArray(arrayName string, schema *jsonschema.Schema, output *strings.Builder, isOptional bool, writeComments bool, extractor *InstanceValueExtractor, propertyPath string) error {
	itemSchema := schema.Items

	if itemSchema == nil {
		if isOptional {
			output.WriteString("# ")
		}
		fmt.Fprintf(output, "%s = []\n", arrayName)
		return nil
	}

	commentPrefix := ""
	if isOptional {
		commentPrefix = "# "
	}

	// Try to get array values from instance
	if extractor != nil {
		if arrayValue, itemType, exists := extractor.GetArrayValue(propertyPath); exists && arrayValue.Len() > 0 {
			return g.formatInstanceArray(arrayName, arrayValue, itemType, itemSchema, output, isOptional, writeComments, extractor, propertyPath)
		}
	}

	switch itemSchema.Type {
	case "object":
		// TOML array of objects syntax: [[table_name]]
		fmt.Fprintf(output, "%s[[%s]]\n", commentPrefix, arrayName)

		if itemSchema.Properties != nil && itemSchema.Properties.Len() > 0 {
			return g.recursivelyEncode(itemSchema, nil, output, arrayName, isOptional, writeComments, false, extractor)
		}
	case "array":
		// nested array - rare but possible
		// TOML doesn't handle nested arrays well, so we simplify
		// we dont have this case in our app config as of now, will handle this seprately later on if needed
		if isOptional {
			output.WriteString("# ")
		}
		fmt.Fprintf(output, "%s = [[]]\n", arrayName)
	default:
		defaultValue := generateDefaultByType(itemSchema.Type)
		arrayValue := fmt.Sprintf("[%s]", defaultValue)

		if isOptional {
			output.WriteString("# ")
		}
		fmt.Fprintf(output, "%s = %s\n", arrayName, arrayValue)
	}

	return nil
}

func (g *ConfigGen) generateFieldLine(fieldName string, schema *jsonschema.Schema, extractor *InstanceValueExtractor, propertyPath string) string {
	var defaultValue string

	// try to get value from instance
	if extractor != nil {
		if instanceValue, exists := extractor.GetFieldValue(propertyPath); exists {
			defaultValue = formatTOMLValue(instanceValue, schema.Type)
			return fmt.Sprintf("%s = %s", fieldName, defaultValue)
		}
	}

	// if not present in instance, use schema default if EnableDefaults is true
	if g.EnableDefaults && schema.Default != nil {
		defaultValue = formatTOMLValue(schema.Default, schema.Type)
		return fmt.Sprintf("%s = %s", fieldName, defaultValue)
	}

	// if default not present in schema, fall back to type-based defaults
	defaultValue = generateDefaultByType(schema.Type)
	return fmt.Sprintf("%s = %s", fieldName, defaultValue)
}

// formatTOMLValue formats a value for TOML based on its type
func formatTOMLValue(value any, schemaType string) string {
	if value == nil {
		return generateDefaultByType(schemaType)
	}

	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v)
	case bool:
		return fmt.Sprintf("%t", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	default:
		// For complex types, convert to string representation
		return fmt.Sprintf(`"%v"`, v)
	}
}

// generateDefaultByType returns a placeholder default value based on JSON schema type
func generateDefaultByType(schemaType string) string {
	switch schemaType {
	case "string":
		return `""`
	case "number", "integer":
		return "0"
	case "boolean":
		return "false"
	case "array":
		return "[]"
	case "object":
		return "{}"
	default:
		return `""`
	}
}

// formatInstanceArray formats an array value from an instance for TOML output
func (g *ConfigGen) formatInstanceArray(arrayName string, arrayValue reflect.Value, itemType string, itemSchema *jsonschema.Schema, output *strings.Builder, isOptional bool, writeComments bool, extractor *InstanceValueExtractor, propertyPath string) error {
	commentPrefix := ""
	if isOptional {
		commentPrefix = "# "
	}

	if itemType == "object" {
		// print array headers using toml table
		for i := 0; i < arrayValue.Len(); i++ {
			fmt.Fprintf(output, "%s[[%s]]\n", commentPrefix, arrayName)

			item := arrayValue.Index(i)
			if item.Kind() == reflect.Ptr {
				if !item.IsNil() {
					item = item.Elem()
				}
			}

			if item.IsValid() && !item.IsZero() {
				itemExtractor := NewInstanceValueExtractor(item.Interface())
				if itemSchema.Properties != nil && itemSchema.Properties.Len() > 0 {
					// use empty prefix "" for array items since itemExtractor is scoped to the item
					// use skipNonRequired=false to ensure all fields from instance are included
					err := g.recursivelyEncode(itemSchema, nil, output, "", isOptional, false, false, itemExtractor)
					if err != nil {
						return err
					}
				}
			}

			output.WriteString("\n")
		}
	} else {
		var items []string
		for i := 0; i < arrayValue.Len(); i++ {
			item := arrayValue.Index(i)
			if item.Kind() == reflect.Pointer && !item.IsNil() {
				item = item.Elem()
			}
			if item.IsValid() {
				formatted := formatTOMLValue(item.Interface(), itemSchema.Type)
				items = append(items, formatted)
			}
		}

		fmt.Fprintf(output, "%s%s = [%s]\n", commentPrefix, arrayName, strings.Join(items, ", "))
	}

	return nil
}
