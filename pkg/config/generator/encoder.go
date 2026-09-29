package generator

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/invopop/jsonschema"
)

func extractPropertyName(path string) string {
	if !strings.Contains(path, ".") {
		return path
	}
	lastDotIndex := strings.LastIndex(path, ".")
	return path[lastDotIndex+1:]
}

type encodePhase int

const (
	phaseAll encodePhase = iota
	phaseLines
	phaseBlocks
)

func rendersBlock(s *jsonschema.Schema) bool {
	switch s.Type {
	case "object":
		return true
	case "array":
		if s.Items != nil && (s.Items.Type == "object" ||
			(s.Items.Items != nil && s.Items.Items.Type == "object")) {
			return true
		}
		return false
	default:
		return false
	}
}

func (g *ConfigGen) recursivelyEncode(schema *jsonschema.Schema, output *strings.Builder, prefix string, parentOptional bool, writeComments bool, skipNonRequired bool, extractor *InstanceValueExtractor, phase encodePhase, instanceOnly bool) error {
	if schema == nil || schema.Properties == nil {
		return fmt.Errorf("schema or properties is nil")
	}

	skipNonRequired = skipNonRequired || g.SkipNonRequired

	requiredFields := make(map[string]bool)
	for _, fieldName := range schema.Required {
		requiredFields[fieldName] = true
	}

	oneOfMembers, oneOfChosen := selectOneOfBranch(schema, extractor, prefix)

	for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		propertyName := pair.Key
		propertySchema := pair.Value

		if slices.Contains(IgnoredProperties, propertyName) {
			continue
		}

		isRequired := requiredFields[propertyName]
		isOneOfMember := oneOfMembers[propertyName]
		isOneOfAlternative := isOneOfMember && !oneOfChosen[propertyName]
		if oneOfChosen[propertyName] {
			isRequired = true
		}

		fullPath := propertyName
		if prefix != "" {
			fullPath = prefix + "." + propertyName
		}

		hasInstanceValue := false
		if extractor != nil {
			hasInstanceValue = extractor.HasValue(fullPath)
			if !hasInstanceValue && strings.Contains(fullPath, ".") {
				simpleName := extractPropertyName(fullPath)
				hasInstanceValue = extractor.HasValue(simpleName)
			}
		}

		if instanceOnly && !hasInstanceValue {
			continue
		}

		isOptional := skipNonRequired && (!isRequired || parentOptional) && !hasInstanceValue

		if isOneOfAlternative {
			isOptional = true
		} else if skipNonRequired && !isRequired && !hasInstanceValue {
			continue
		}

		if !g.EnableDeprecated && propertySchema.Deprecated {
			continue
		}

		if phase != phaseAll {
			block := rendersBlock(propertySchema)
			if (phase == phaseLines) == block {
				continue
			}
		}

		if writeComments && g.EnableInfoComments {
			g.writePropertyComments(propertySchema, output)
		}

		// why: A typeless property backed by a oneOf (commonly array|null) must be
		// rendered as its real type, not as an empty string.
		propertyType := propertySchema.Type
		if propertyType == "" && oneOfIsArray(propertySchema) {
			if isOptional {
				output.WriteString("# ")
			}
			fmt.Fprintf(output, "%s = []\n", fullPath)
			if writeComments && g.EnableInfoComments {
				output.WriteString("\n")
			}
			continue
		}

		switch propertyType {
		case "array":
			err := g.encodeTOMLArray(fullPath, propertySchema, output, isOptional, writeComments, extractor, fullPath, instanceOnly)
			if err != nil {
				return err
			}
			output.WriteString("\n")
		case "object":
			err := g.encodeTOMLObject(fullPath, propertySchema, output, isOptional, writeComments, extractor, fullPath, instanceOnly)
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
	}
	return nil
}

func selectOneOfBranch(schema *jsonschema.Schema, extractor *InstanceValueExtractor, prefix string) (members, chosen map[string]bool) {
	members = map[string]bool{}
	chosen = map[string]bool{}

	chosenIdx := -1
	for i, branch := range schema.OneOf {
		if len(branch.Required) == 0 {
			continue
		}
		if chosenIdx == -1 {
			chosenIdx = i
		}
		for _, field := range branch.Required {
			members[field] = true
			if extractor == nil {
				continue
			}
			fullPath := field
			if prefix != "" {
				fullPath = prefix + "." + field
			}
			if extractor.HasValue(fullPath) || extractor.HasValue(field) {
				chosenIdx = i
			}
		}
	}

	if chosenIdx >= 0 {
		for _, field := range schema.OneOf[chosenIdx].Required {
			chosen[field] = true
		}
	}

	return members, chosen
}

func oneOfIsArray(schema *jsonschema.Schema) bool {
	for _, branch := range schema.OneOf {
		if branch.Type == "array" {
			return true
		}
	}
	return false
}

func (g *ConfigGen) writePropertyComments(schema *jsonschema.Schema, output *strings.Builder) {
	if schema.Title != "" {
		output.WriteString("# ")
		output.WriteString(schema.Title)
		output.WriteString("\n")
	}

	if schema.Description != "" {
		for line := range strings.SplitSeq(schema.Description, "\n") {
			output.WriteString("# ")
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	if len(schema.Examples) > 0 {
		output.WriteString("# Examples: ")
		for i, example := range schema.Examples {
			if i > 0 {
				output.WriteString(", ")
			}
			exampleStr := fmt.Sprintf("%v", example)
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

	if isOptional {
		output.WriteString("# ")
	}

	output.WriteString(fieldLine)
	// TODO(sk): remove this if comments are diabled
	output.WriteString("\n")
}

func (g *ConfigGen) encodeTOMLObject(tableName string, schema *jsonschema.Schema, output *strings.Builder, isOptional bool, writeComments bool, extractor *InstanceValueExtractor, propertyPath string, instanceOnly bool) error {
	if schema.Properties == nil || schema.Properties.Len() == 0 {
		if extractor != nil {
			mapValue, exists := extractor.GetMapValue(propertyPath)
			if !exists && strings.Contains(propertyPath, ".") {
				simpleName := extractPropertyName(propertyPath)
				mapValue, exists = extractor.GetMapValue(simpleName)
			}

			if exists && len(mapValue) > 0 {
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

	nestedExtractor := extractor
	if extractor != nil {
		nestedValue, exists := extractor.GetFieldValue(propertyPath)
		if !exists && strings.Contains(propertyPath, ".") {
			simpleName := extractPropertyName(propertyPath)
			nestedValue, exists = extractor.GetFieldValue(simpleName)
		}

		if exists {
			nestedExtractor = NewInstanceValueExtractor(nestedValue)
		}
	}

	g.recursivelyEncode(schema, output, tableName, isOptional, writeComments, false, nestedExtractor, phaseAll, instanceOnly)
	return nil
}

func (g *ConfigGen) encodeTOMLArray(arrayName string, schema *jsonschema.Schema, output *strings.Builder, isOptional bool, writeComments bool, extractor *InstanceValueExtractor, propertyPath string, instanceOnly bool) error {
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

	if extractor != nil {
		arrayValue, itemType, exists := extractor.GetArrayValue(propertyPath)
		if !exists && strings.Contains(propertyPath, ".") {
			simpleName := extractPropertyName(propertyPath)
			arrayValue, itemType, exists = extractor.GetArrayValue(simpleName)
		}
		if exists && arrayValue.Len() > 0 {
			return g.formatInstanceArray(arrayName, arrayValue, itemType, itemSchema, output, isOptional, writeComments, extractor, propertyPath)
		}
	}

	switch itemSchema.Type {
	case "object":
		fmt.Fprintf(output, "%s[[%s]]\n", commentPrefix, arrayName)

		if itemSchema.Properties != nil && itemSchema.Properties.Len() > 0 {
			return g.recursivelyEncode(itemSchema, output, arrayName, isOptional, writeComments, false, extractor, phaseAll, instanceOnly)
		}
	case "array":
		if itemSchema.Items != nil && itemSchema.Items.Type == "object" {
			fmt.Fprintf(output, "%s[[%s]]\n", commentPrefix, arrayName)
			if itemSchema.Items.Properties != nil && itemSchema.Items.Properties.Len() > 0 {
				return g.recursivelyEncode(itemSchema.Items, output, arrayName, isOptional, writeComments, false, extractor, phaseAll, instanceOnly)
			}
		} else {
			if isOptional {
				output.WriteString("# ")
			}
			fmt.Fprintf(output, "%s = []\n", arrayName)
		}
	default:
		if isOptional {
			output.WriteString("# ")
		}
		fmt.Fprintf(output, "%s = []\n", arrayName)
	}

	return nil
}

func (g *ConfigGen) generateFieldLine(fieldName string, schema *jsonschema.Schema, extractor *InstanceValueExtractor, propertyPath string) string {
	var defaultValue string

	if extractor != nil {
		if instanceValue, exists := extractor.GetFieldValue(propertyPath); exists {
			defaultValue = formatTOMLValue(instanceValue, schema.Type)
			return fmt.Sprintf("%s = %s", fieldName, defaultValue)
		}
		if strings.Contains(propertyPath, ".") {
			simpleName := extractPropertyName(propertyPath)
			if instanceValue, exists := extractor.GetFieldValue(simpleName); exists {
				defaultValue = formatTOMLValue(instanceValue, schema.Type)
				return fmt.Sprintf("%s = %s", fieldName, defaultValue)
			}
		}
	}

	if g.EnableDefaults && schema.Default != nil {
		defaultValue = formatTOMLValue(schema.Default, schema.Type)
		return fmt.Sprintf("%s = %s", fieldName, defaultValue)
	}

	defaultValue = generateDefaultByType(schema.Type)
	return fmt.Sprintf("%s = %s", fieldName, defaultValue)
}

func formatTOMLValue(value any, schemaType string) string {
	if value == nil {
		return generateDefaultByType(schemaType)
	}

	switch v := value.(type) {
	case string:
		formattedString := ""
		if strings.Contains(v, "\n") {
			formattedString = fmt.Sprintf("\"\"\"\n%s\"\"\"", v)
		} else {
			formattedString = fmt.Sprintf(`"%s"`, v)
		}
		return formattedString
	case bool:
		return fmt.Sprintf("%t", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	default:
		return fmt.Sprintf(`"%v"`, v)
	}
}

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

func (g *ConfigGen) formatInstanceArray(arrayName string, arrayValue reflect.Value, itemType string, itemSchema *jsonschema.Schema, output *strings.Builder, isOptional bool, writeComments bool, extractor *InstanceValueExtractor, propertyPath string) error {
	commentPrefix := ""
	if isOptional {
		commentPrefix = "# "
	}

	if itemType == "object" {
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
					err := g.recursivelyEncode(itemSchema, output, arrayName, isOptional, false, false, itemExtractor, phaseAll, true)
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

			if !item.IsValid() || item.IsZero() {
				continue
			}

			if item.Kind() == reflect.String && item.String() == "" {
				continue
			}

			formatted := formatTOMLValue(item.Interface(), itemSchema.Type)
			items = append(items, formatted)
		}

		if len(items) > 0 {
			fmt.Fprintf(output, "%s%s = [%s]\n", commentPrefix, arrayName, strings.Join(items, ", "))
		} else {
			fmt.Fprintf(output, "%s%s = []\n", commentPrefix, arrayName)
		}
	}

	return nil
}
