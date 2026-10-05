package diagnostics

import (
	"bufio"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/config/schema"
	tomlparser "github.com/nuonco/nuon/pkg/parser/toml"
)

const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// Diagnostic is a per-file schema finding. Line and character are zero-based,
// matching the TOML parser and the language server.
type Diagnostic struct {
	Severity  string
	Message   string
	StartLine int
	StartChar int
	EndLine   int
	EndChar   int
}

func at(severity, message string, r tomlparser.Range) Diagnostic {
	return Diagnostic{
		Severity:  severity,
		Message:   message,
		StartLine: r.Start.Line,
		StartChar: r.Start.Character,
		EndLine:   r.End.Line,
		EndChar:   r.End.Character,
	}
}

func atLine(severity, message string) Diagnostic {
	return Diagnostic{Severity: severity, Message: message}
}

// DetectSchemaType returns the schema type from the first comment line.
func DetectSchemaType(text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if comment != "" {
				return comment
			}
		}
		break
	}
	return ""
}

// DetectSchemaTypeForDocument uses the leading comment, then branch.toml.
func DetectSchemaTypeForDocument(text, uri string) string {
	if schemaType := DetectSchemaType(text); schemaType != "" {
		return schemaType
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	if strings.EqualFold(path.Base(parsed.Path), "branch.toml") {
		return "branch"
	}
	return ""
}

// DiagnoseFile runs the same per-file checks as the language server.
// A file with no schema comment returns no diagnostics.
func DiagnoseFile(filename, text string) []Diagnostic {
	schemaType := DetectSchemaTypeForDocument(text, filename)
	if schemaType == "" {
		return nil
	}
	if !schema.IsValidSchemaType(schemaType) {
		types := schema.GetSchemaTypes()
		sort.Strings(types)
		return []Diagnostic{{
			Severity: SeverityError,
			Message:  "Unknown schema type '" + schemaType + "'. Valid types: " + strings.Join(types, ", "),
		}}
	}
	root, err := schema.LookupSchemaType(schemaType)
	if err != nil || root == nil {
		return nil
	}
	doc := tomlparser.ParseToml(text)
	doc.Values = Values(text, doc)
	return Diagnose(doc, root)
}

// Values fills parsed values, falling back to raw text when strict parse fails.
func Values(text string, doc *tomlparser.TomlDocument) map[string]any {
	if strictDoc, err := tomlparser.ParseStrict(text); err == nil {
		return strictDoc.Values
	}
	return ExtractRawValues(text, doc)
}

// DiagnoseDocument generates diagnostics for a TOML document
func Diagnose(doc *tomlparser.TomlDocument, rootSchema *jsonschema.Schema) []Diagnostic {
	diagnostics := []Diagnostic{}

	// Defensive guard: return empty diagnostics if rootSchema is nil
	if rootSchema == nil {
		return diagnostics
	}

	defs := make(map[string]*jsonschema.Schema)
	if rootSchema.Definitions != nil {
		defs = rootSchema.Definitions
	}

	effectiveRoot := mergeAllOf(rootSchema, defs)

	// 1. Unknown keys
	for _, key := range doc.Keys {
		// Defensive guard: skip keys with empty Path
		if len(key.Path) == 0 {
			continue
		}

		// Determine the table path for this key
		parentPath := key.Path[:len(key.Path)-1]

		// Resolve parent schema
		parentSchema := ResolveSchema(effectiveRoot, parentPath, defs)
		if parentSchema == nil {
			continue
		}

		// Check if property exists
		found := false
		if parentSchema.Properties != nil && parentSchema.Properties.Len() > 0 {
			_, found = parentSchema.Properties.Get(key.Name)
		} else {
			// If properties are empty/nil, it's likely a map (additionalProperties) or allows everything
			// We skip "Unknown key" check unless we can verify additionalProperties is false
			found = true
		}

		if !found {
			diagnostics = append(diagnostics, at(SeverityWarning, fmt.Sprintf("Unknown key: %s", key.Name), key.Range))
		} else {
			// 3. Type mismatches (only if key is known)
			fullPath := strings.Join(key.Path, ".")
			if val, ok := doc.Values[fullPath]; ok {
				var propSchema *jsonschema.Schema
				if parentSchema.Properties != nil {
					if ps, ok := parentSchema.Properties.Get(key.Name); ok {
						propSchema = ps
					}
				}
				// If not in properties, use AdditionalProperties (e.g. for maps)
				if propSchema == nil {
					propSchema = parentSchema.AdditionalProperties
				}

				if propSchema != nil {
					// Resolve ref for property schema if needed for type checking
					if propSchema.Ref != "" {
						if r := resolveRef(propSchema.Ref, defs); r != nil {
							propSchema = r
						}
					}

					if !schemaTypeMatches(propSchema, TomlTypeOf(val)) {
						diagnostics = append(diagnostics, at(SeverityError, fmt.Sprintf("Type mismatch for '%s': expected %s, got %s", key.Name, propSchema.Type, TomlTypeOf(val)), key.Range))
					} else if !enumAllows(propSchema, val) {
						diagnostics = append(diagnostics, at(SeverityError, fmt.Sprintf("Invalid value for '%s': %v. Valid values: %s", key.Name, val, formatEnum(propSchema.Enum)), key.Range))
					}
				}
			}
		}
	}

	// 2. Missing required fields
	for _, table := range doc.Tables {
		schemaNode := ResolveSchema(effectiveRoot, table.Path, defs)
		if schemaNode == nil {
			continue
		}

		for _, reqField := range schemaNode.Required {
			if !PropertyExists(doc, table.Path, reqField) {
				diagnostics = append(diagnostics, at(SeverityError, fmt.Sprintf("Missing required field '%s' for table %s", reqField, table.Name), table.Range))
			}
		}
	}

	// Also check required fields for the root table (empty path)
	// Use effectiveRoot for root checks
	if effectiveRoot != nil {
		for _, reqField := range effectiveRoot.Required {
			if !PropertyExists(doc, []string{}, reqField) {
				diagnostics = append(diagnostics, atLine(SeverityError, fmt.Sprintf("Missing required field '%s' for root table", reqField)))
			}
		}
	}

	// 4. oneOf validation
	if len(effectiveRoot.OneOf) > 0 {
		satisfiedCount := 0
		var satisfiedTitles []string
		var satisfiedFields []string
		var requiredLists []string

		for _, branch := range effectiveRoot.OneOf {
			if branch == nil {
				continue
			}
			allPresent := true
			if branch.Properties != nil && branch.Properties.Len() > 0 && len(branch.Required) == 0 {
				continue
			}

			for _, req := range branch.Required {
				if !PropertyExists(doc, []string{}, req) {
					allPresent = false
					break
				}
			}

			if allPresent && len(branch.Required) > 0 {
				satisfiedCount++
				if branch.Title != "" {
					satisfiedTitles = append(satisfiedTitles, branch.Title)
				} else {
					satisfiedTitles = append(satisfiedTitles, fmt.Sprintf("[%s]", strings.Join(branch.Required, ", ")))
				}
				// Track fields that satisfied this branch
				satisfiedFields = append(satisfiedFields, branch.Required...)
			}

			if len(branch.Required) > 0 {
				if branch.Title != "" {
					requiredLists = append(requiredLists, branch.Title)
				} else {
					requiredLists = append(requiredLists, fmt.Sprintf("[%s]", strings.Join(branch.Required, ", ")))
				}
			}
		}

		if satisfiedCount == 0 {
			diagnostics = append(diagnostics, atLine(SeverityError, fmt.Sprintf("This document must contain exactly one of: %s", strings.Join(requiredLists, ", "))))
		} else if satisfiedCount > 1 {
			// Multiple oneOf branches satisfied - underline all conflicting fields
			foundAny := false

			for _, fieldName := range satisfiedFields {
				// Find the key or table for this field at root level (empty tablePath)
				found := false

				// Check for root-level keys (e.g., "name = value")
				for _, key := range doc.Keys {
					if len(key.Path) == 1 && key.Name == fieldName {
						diagnostics = append(diagnostics, at(SeverityError, fmt.Sprintf("Only one of: %s may be defined; found %s", strings.Join(requiredLists, ", "), strings.Join(satisfiedTitles, ", ")), key.Range))
						found = true
						foundAny = true
						break
					}
				}

				// Also check for tables (e.g., [connected_repo], [public_repo])
				if !found {
					for _, table := range doc.Tables {
						if len(table.Path) == 1 && table.Name == fieldName {
							diagnostics = append(diagnostics, at(SeverityError, fmt.Sprintf("Only one of: %s may be defined; found %s", strings.Join(requiredLists, ", "), strings.Join(satisfiedTitles, ", ")), table.Range))
							found = true
							foundAny = true
							break
						}
					}
				}
			}
			// If we didn't find any fields to underline, fall back to line 0
			if !foundAny {
				diagnostics = append(diagnostics, atLine(SeverityError, fmt.Sprintf("Only one of: %s may be defined; found %s", strings.Join(requiredLists, ", "), strings.Join(satisfiedTitles, ", "))))
			}
		}
	}

	return diagnostics
}

// Helper functions

func mergeAllOf(schema *jsonschema.Schema, defs map[string]*jsonschema.Schema) *jsonschema.Schema {
	if schema == nil {
		return nil
	}

	current := schema
	if current.Ref != "" {
		if r := resolveRef(current.Ref, defs); r != nil {
			current = r
		}
	}

	if len(current.AllOf) == 0 {
		return current
	}

	merged := &jsonschema.Schema{
		Type:       current.Type,
		Properties: current.Properties,
		Required:   append([]string{}, current.Required...),
		OneOf:      current.OneOf,
	}
	if merged.Properties == nil {
		merged.Properties = jsonschema.NewProperties()
	}

	for _, branch := range current.AllOf {
		if branch == nil {
			continue
		}

		// Merge definitions first so $ref resolution works for this branch
		for k, v := range branch.Definitions {
			if _, exists := defs[k]; !exists {
				defs[k] = v
			}
		}

		resolved := branch
		if branch.Ref != "" {
			if r := resolveRef(branch.Ref, defs); r != nil {
				resolved = r
			}
		}

		if resolved.Properties != nil {
			pair := resolved.Properties.Oldest()
			for pair != nil {
				if _, exists := merged.Properties.Get(pair.Key); !exists {
					merged.Properties.Set(pair.Key, pair.Value)
				}
				pair = pair.Next()
			}
		}

		merged.Required = append(merged.Required, resolved.Required...)

		if len(resolved.OneOf) > 0 {
			merged.OneOf = append(merged.OneOf, resolved.OneOf...)
		}

		// Also merge definitions from the resolved schema (in case ref target has its own)
		for k, v := range resolved.Definitions {
			if _, exists := defs[k]; !exists {
				defs[k] = v
			}
		}
	}

	return merged
}

func ResolveSchema(root *jsonschema.Schema, path []string, defs map[string]*jsonschema.Schema) *jsonschema.Schema {
	current := root
	// Resolve root ref/array
	if current.Ref != "" {
		if r := resolveRef(current.Ref, defs); r != nil {
			current = r
		}
	}

	for _, segment := range path {
		if current == nil {
			return nil
		}

		// If array, peel off Items
		if current.Type == "array" && current.Items != nil {
			current = current.Items
			if current.Ref != "" {
				if r := resolveRef(current.Ref, defs); r != nil {
					current = r
				}
			}
		}

		if current.Properties == nil {
			return nil
		}

		prop, ok := current.Properties.Get(segment)
		if !ok {
			return nil
		}

		current = prop
		if current == nil {
			return nil
		}
		if current.Ref != "" {
			if r := resolveRef(current.Ref, defs); r != nil {
				current = r
			}
		}
	}

	// If we ended on an array (e.g. path pointed to a table which is an array of tables),
	// we usually want the ITEM schema to check keys against.
	if current != nil && current.Type == "array" && current.Items != nil {
		current = current.Items
		if current.Ref != "" {
			if r := resolveRef(current.Ref, defs); r != nil {
				current = r
			}
		}
	}

	return current
}

// resolveRef resolves a JSON Schema $ref to its definition
func resolveRef(ref string, defsLookup map[string]*jsonschema.Schema) *jsonschema.Schema {
	ref = strings.TrimPrefix(ref, "#/definitions/")
	ref = strings.TrimPrefix(ref, "#/$defs/")
	return defsLookup[ref]
}

func PropertyExists(doc *tomlparser.TomlDocument, tablePath []string, keyName string) bool {
	// Check keys
	if keyExistsInKeys(doc, tablePath, keyName) {
		return true
	}

	// Check tables
	// A required property might be satisfied by a table (e.g. [public_repo])
	targetPathLen := len(tablePath) + 1
	for _, t := range doc.Tables {
		if len(t.Path) != targetPathLen {
			continue
		}
		// Last segment of path must match keyName
		if t.Path[len(t.Path)-1] != keyName {
			continue
		}

		// Prefix must match tablePath
		match := true
		for i, p := range tablePath {
			if t.Path[i] != p {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}

func keyExistsInKeys(doc *tomlparser.TomlDocument, tablePath []string, keyName string) bool {
	targetLen := len(tablePath) + 1

	for _, k := range doc.Keys {
		if len(k.Path) != targetLen {
			continue
		}
		if k.Name != keyName {
			continue
		}

		match := true
		for i, p := range tablePath {
			if k.Path[i] != p {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TomlTypeOf(value any) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int64, int32:
		return "integer"
	case float64, float32:
		return "number"
	case bool:
		return "boolean"
	case []any, []string, []map[string]any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return "unknown"
	}
}

func schemaTypeMatches(schemaNode *jsonschema.Schema, gotType string) bool {
	if schemaNode == nil {
		return false
	}
	if schemaNode.Type == "" {
		return true
	}
	if schemaNode.Type == gotType {
		return true
	}
	if schemaNode.Type == "number" && gotType == "integer" {
		return true
	}
	return false
}

func enumAllows(schemaNode *jsonschema.Schema, value any) bool {
	if schemaNode == nil || len(schemaNode.Enum) == 0 {
		return true
	}

	switch value.(type) {
	case string, int, int64, int32, float64, float32, bool:
	default:
		return true
	}

	got := normalizeEnumValue(value)
	for _, allowed := range schemaNode.Enum {
		if normalizeEnumValue(allowed) == got {
			return true
		}
	}
	return false
}

// normalizeEnumValue strips the quotes that extractRawValues leaves on string
// values when a document fails to parse strictly.
func normalizeEnumValue(value any) string {
	s := fmt.Sprintf("%v", value)
	if len(s) >= 2 {
		if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
			(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
			s = s[1 : len(s)-1]
		}
	}
	return s
}

func formatEnum(enum []any) string {
	values := make([]string, 0, len(enum))
	for _, val := range enum {
		values = append(values, fmt.Sprintf("%v", val))
	}
	return strings.Join(values, ", ")
}

// extractRawValues extracts raw value strings from TOML text when strict parsing fails
// This allows type checking even for invalid TOML syntax like "terraform_version = 1.2.3"
func ExtractRawValues(text string, doc *tomlparser.TomlDocument) map[string]any {
	values := make(map[string]any)
	lines := strings.Split(text, "\n")

	for _, key := range doc.Keys {
		if key.Range.Start.Line >= len(lines) {
			continue
		}

		line := lines[key.Range.Start.Line]

		// Find the equals sign
		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			continue
		}

		// Extract value part (after =)
		rawValue := strings.TrimSpace(line[eqIdx+1:])
		if rawValue == "" {
			continue
		}

		// Try to parse as number for type checking purposes
		// This helps catch cases like "1.2.3" which are invalid floats
		var value any = rawValue

		switch {
		case strings.HasPrefix(rawValue, "["):
			value = []any{}
		case strings.HasPrefix(rawValue, "{"):
			value = map[string]any{}
		case rawValue == "true" || rawValue == "false":
			value = rawValue == "true"
		default:
			// Try parsing as number
			if f, err := strconv.ParseFloat(rawValue, 64); err == nil {
				// Check if it looks like a float or integer
				if strings.Contains(rawValue, ".") {
					value = f
				} else {
					value = int64(f)
				}
			} else if looksLikeNumber(rawValue) {
				// For values that failed to parse as floats, check if they look like
				// they were attempting to be numbers (e.g., "1.2.3" which has multiple dots)
				// Treat as a float64 for type mismatch detection
				value = 0.0
			}
		}

		fullPath := strings.Join(key.Path, ".")
		values[fullPath] = value
	}

	return values
}

// looksLikeNumber returns true if a string appears to be a numeric value
// that failed to parse (e.g., "1.2.3" with multiple decimal points)
func looksLikeNumber(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Remove quotes if present
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return false
	}

	// Check if it starts with a digit or negative sign
	firstChar := s[0]
	if (firstChar < '0' || firstChar > '9') && firstChar != '-' && firstChar != '+' {
		return false
	}

	// Check if all characters are digits, dots, or negative sign
	dotCount := 0
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			continue
		}
		if ch == '.' {
			dotCount++
			continue
		}
		if ch == '-' || ch == '+' {
			continue
		}
		if ch == 'e' || ch == 'E' {
			// Scientific notation is valid TOML float
			continue
		}
		return false
	}

	// If it has multiple dots or ends with a dot, it's malformed numeric
	return dotCount > 1 || strings.HasSuffix(s, ".")
}
