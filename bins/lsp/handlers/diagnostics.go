package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon/bins/lsp/models"
	tomlparser "github.com/nuonco/nuon/pkg/parser/toml"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// PublishDiagnostics handles the full diagnostic cycle: detection, parsing, diagnosis, and publishing
func PublishDiagnostics(ctx *glsp.Context, uri protocol.DocumentUri, text string) {
	// Detect schema type
	schemaType := models.DetectSchemaType(text)
	if schemaType == "" {
		// Clear diagnostics if no schema detected
		ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: []protocol.Diagnostic{},
		})
		return
	}

	// Lookup schema
	schema, err := models.LookupSchema(schemaType)
	if err != nil {
		log.Errorf("❌ Schema lookup error during diagnostics: %v", err)
		ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: []protocol.Diagnostic{},
		})
		return
	}
	if schema == nil {
		log.Warningf("⚠️  No schema found for type '%s' during diagnostics", schemaType)
		ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: []protocol.Diagnostic{},
		})
		return
	}

	// Parse TOML (always succeeds with loose parser)
	doc := tomlparser.ParseToml(text)

	// Attempt strict parse to get values for type checking
	// If strict parse fails (invalid syntax), extract values from raw TOML text
	if strictDoc, err := tomlparser.ParseStrict(text); err == nil {
		doc.Values = strictDoc.Values
	} else {
		// Strict parsing failed - extract raw value strings from text for type analysis
		doc.Values = extractRawValues(text, doc)
	}

	// Generate diagnostics
	diags := DiagnoseDocument(uri, doc, schema)
	log.Infof("🩺 Generated %d diagnostics for %s", len(diags), uri)

	// Publish
	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diags,
	})
}

// DiagnoseDocument generates diagnostics for a TOML document
func DiagnoseDocument(uri protocol.DocumentUri, doc *tomlparser.TomlDocument, rootSchema *jsonschema.Schema) []protocol.Diagnostic {
	diagnostics := []protocol.Diagnostic{}

	// Defensive guard: return empty diagnostics if rootSchema is nil
	if rootSchema == nil {
		return diagnostics
	}

	defs := make(map[string]*jsonschema.Schema)
	if rootSchema.Definitions != nil {
		defs = rootSchema.Definitions
	}

	effectiveRoot := rootSchema
	if effectiveRoot.Ref != "" {
		if r := resolveRef(effectiveRoot.Ref, defs); r != nil {
			effectiveRoot = r
		}
	}

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
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Severity: ptrSeverity(protocol.DiagnosticSeverityWarning),
				Message:  fmt.Sprintf("Unknown key: %s", key.Name),
				Range:    toProtocolRange(key.Range),
				Source:   ptr("Nuon LSP"),
			})
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

					if !schemaTypeMatches(propSchema, tomlTypeOf(val)) {
						diagnostics = append(diagnostics, protocol.Diagnostic{
							Severity: ptrSeverity(protocol.DiagnosticSeverityError),
							Message:  fmt.Sprintf("Type mismatch for '%s': expected %s, got %s", key.Name, propSchema.Type, tomlTypeOf(val)),
							Range:    toProtocolRange(key.Range),
							Source:   ptr("Nuon LSP"),
						})
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
				diagnostics = append(diagnostics, protocol.Diagnostic{
					Severity: ptrSeverity(protocol.DiagnosticSeverityError),
					Message:  fmt.Sprintf("Missing required field '%s' for table %s", reqField, table.Name),
					Range:    toProtocolRange(table.Range),
					Source:   ptr("Nuon LSP"),
				})
			}
		}
	}

	// Also check required fields for the root table (empty path)
	// Use effectiveRoot for root checks
	if effectiveRoot != nil {
		for _, reqField := range effectiveRoot.Required {
			if !PropertyExists(doc, []string{}, reqField) {
				diagnostics = append(diagnostics, protocol.Diagnostic{
					Severity: ptrSeverity(protocol.DiagnosticSeverityError),
					Message:  fmt.Sprintf("Missing required field '%s' for root table", reqField),
					Range:    protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 0, Character: 0}},
					Source:   ptr("Nuon LSP"),
				})
			}
		}
	}

	// 4. oneOf validation
	if len(effectiveRoot.OneOf) > 0 {
		satisfiedCount := 0
		var satisfiedTitles []string
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
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Severity: ptrSeverity(protocol.DiagnosticSeverityError),
				Message:  fmt.Sprintf("This document must contain exactly one of: %s", strings.Join(requiredLists, ", ")),
				Range:    protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 0, Character: 0}},
				Source:   ptr("Nuon LSP"),
			})
		} else if satisfiedCount > 1 {
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Severity: ptrSeverity(protocol.DiagnosticSeverityError),
				Message:  fmt.Sprintf("Only one of: %s may be defined; found %s", strings.Join(requiredLists, ", "), strings.Join(satisfiedTitles, ", ")),
				Range:    protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 0, Character: 0}},
				Source:   ptr("Nuon LSP"),
			})
		}
	}

	return diagnostics
}

// Helper functions

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

func tomlTypeOf(value any) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int64, int32:
		return "integer"
	case float64, float32:
		return "number"
	case bool:
		return "boolean"
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

func toProtocolRange(r tomlparser.Range) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{
			Line:      uint32(r.Start.Line),
			Character: uint32(r.Start.Character),
		},
		End: protocol.Position{
			Line:      uint32(r.End.Line),
			Character: uint32(r.End.Character),
		},
	}
}

func ptrSeverity(s protocol.DiagnosticSeverity) *protocol.DiagnosticSeverity {
	return &s
}

// extractRawValues extracts raw value strings from TOML text when strict parsing fails
// This allows type checking even for invalid TOML syntax like "terraform_version = 1.2.3"
func extractRawValues(text string, doc *tomlparser.TomlDocument) map[string]any {
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

		// Try parsing as number
		if f, err := strconv.ParseFloat(rawValue, 64); err == nil {
			// Check if it looks like a float or integer
			if strings.Contains(rawValue, ".") {
				value = f
			} else {
				value = int64(f)
			}
		} else {
			// For values that failed to parse as floats, check if they look like
			// they were attempting to be numbers (e.g., "1.2.3" which has multiple dots)
			// We want to treat these as numeric types for type checking purposes
			if looksLikeNumber(rawValue) {
				// Can't parse as valid float, but looks like a number attempt
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
