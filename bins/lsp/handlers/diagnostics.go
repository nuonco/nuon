package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/nuonco/nuon/bins/lsp/models"
	tomlparser "github.com/nuonco/nuon/pkg/parser/toml"
)

func PublishDiagnostics(ctx *glsp.Context, uri protocol.DocumentUri, text string) {
	var diagnostics []protocol.Diagnostic

	schemaType := models.DetectSchemaTypeForDocument(text, string(uri))
	if schemaType == "" {
		ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: []protocol.Diagnostic{},
		})
		return
	}

	if !models.IsValidSchemaType(schemaType) {
		lines := strings.Split(text, "\n")
		if len(lines) > 0 {
			firstLine := lines[0]
			startChar := strings.Index(firstLine, schemaType)
			if startChar == -1 {
				startChar = 0
			}
			endChar := startChar + len(schemaType)

			validTypes := models.GetValidSchemaTypes()
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Severity: ptrSeverity(protocol.DiagnosticSeverityError),
				Message:  fmt.Sprintf("Unknown schema type '%s'. Valid types: %s", schemaType, strings.Join(validTypes, ", ")),
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: uint32(startChar)},
					End:   protocol.Position{Line: 0, Character: uint32(endChar)},
				},
				Source: ptr("Nuon LSP"),
			})
		}

		ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diagnostics,
		})
		return
	}

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

	doc := tomlparser.ParseToml(text)

	if strictDoc, err := tomlparser.ParseStrict(text); err == nil {
		doc.Values = strictDoc.Values
	} else {
		doc.Values = extractRawValues(text, doc)
	}

	diags := DiagnoseDocument(uri, doc, schema)
	log.Infof("🩺 Generated %d diagnostics for %s", len(diags), uri)

	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diags,
	})
}

func DiagnoseDocument(uri protocol.DocumentUri, doc *tomlparser.TomlDocument, rootSchema *jsonschema.Schema) []protocol.Diagnostic {
	diagnostics := []protocol.Diagnostic{}

	if rootSchema == nil {
		return diagnostics
	}

	defs := make(map[string]*jsonschema.Schema)
	if rootSchema.Definitions != nil {
		defs = rootSchema.Definitions
	}

	effectiveRoot := mergeAllOf(rootSchema, defs)

	for _, key := range doc.Keys {
		if len(key.Path) == 0 {
			continue
		}

		parentPath := key.Path[:len(key.Path)-1]

		parentSchema := ResolveSchema(effectiveRoot, parentPath, defs)
		if parentSchema == nil {
			continue
		}

		found := false
		if parentSchema.Properties != nil && parentSchema.Properties.Len() > 0 {
			_, found = parentSchema.Properties.Get(key.Name)
		} else {
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
			fullPath := strings.Join(key.Path, ".")
			if val, ok := doc.Values[fullPath]; ok {
				var propSchema *jsonschema.Schema
				if parentSchema.Properties != nil {
					if ps, ok := parentSchema.Properties.Get(key.Name); ok {
						propSchema = ps
					}
				}
				if propSchema == nil {
					propSchema = parentSchema.AdditionalProperties
				}

				if propSchema != nil {
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
					} else if !enumAllows(propSchema, val) {
						diagnostics = append(diagnostics, protocol.Diagnostic{
							Severity: ptrSeverity(protocol.DiagnosticSeverityError),
							Message:  fmt.Sprintf("Invalid value for '%s': %v. Valid values: %s", key.Name, val, formatEnum(propSchema.Enum)),
							Range:    toProtocolRange(key.Range),
							Source:   ptr("Nuon LSP"),
						})
					}
				}
			}
		}
	}

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
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Severity: ptrSeverity(protocol.DiagnosticSeverityError),
				Message:  fmt.Sprintf("This document must contain exactly one of: %s", strings.Join(requiredLists, ", ")),
				Range:    protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 0, Character: 0}},
				Source:   ptr("Nuon LSP"),
			})
		} else if satisfiedCount > 1 {
			foundAny := false

			for _, fieldName := range satisfiedFields {
				found := false

				for _, key := range doc.Keys {
					if len(key.Path) == 1 && key.Name == fieldName {
						diagnostics = append(diagnostics, protocol.Diagnostic{
							Severity: ptrSeverity(protocol.DiagnosticSeverityError),
							Message:  fmt.Sprintf("Only one of: %s may be defined; found %s", strings.Join(requiredLists, ", "), strings.Join(satisfiedTitles, ", ")),
							Range:    toProtocolRange(key.Range),
							Source:   ptr("Nuon LSP"),
						})
						found = true
						foundAny = true
						break
					}
				}

				if !found {
					for _, table := range doc.Tables {
						if len(table.Path) == 1 && table.Name == fieldName {
							diagnostics = append(diagnostics, protocol.Diagnostic{
								Severity: ptrSeverity(protocol.DiagnosticSeverityError),
								Message:  fmt.Sprintf("Only one of: %s may be defined; found %s", strings.Join(requiredLists, ", "), strings.Join(satisfiedTitles, ", ")),
								Range:    toProtocolRange(table.Range),
								Source:   ptr("Nuon LSP"),
							})
							found = true
							foundAny = true
							break
						}
					}
				}
			}
			if !foundAny {
				diagnostics = append(diagnostics, protocol.Diagnostic{
					Severity: ptrSeverity(protocol.DiagnosticSeverityError),
					Message:  fmt.Sprintf("Only one of: %s may be defined; found %s", strings.Join(requiredLists, ", "), strings.Join(satisfiedTitles, ", ")),
					Range:    protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: 0, Character: 0}},
					Source:   ptr("Nuon LSP"),
				})
			}
		}
	}

	return diagnostics
}

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
	if current.Ref != "" {
		if r := resolveRef(current.Ref, defs); r != nil {
			current = r
		}
	}

	for _, segment := range path {
		if current == nil {
			return nil
		}

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

func resolveRef(ref string, defsLookup map[string]*jsonschema.Schema) *jsonschema.Schema {
	ref = strings.TrimPrefix(ref, "#/definitions/")
	ref = strings.TrimPrefix(ref, "#/$defs/")
	return defsLookup[ref]
}

func PropertyExists(doc *tomlparser.TomlDocument, tablePath []string, keyName string) bool {
	if keyExistsInKeys(doc, tablePath, keyName) {
		return true
	}

	targetPathLen := len(tablePath) + 1
	for _, t := range doc.Tables {
		if len(t.Path) != targetPathLen {
			continue
		}
		if t.Path[len(t.Path)-1] != keyName {
			continue
		}

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

func extractRawValues(text string, doc *tomlparser.TomlDocument) map[string]any {
	values := make(map[string]any)
	lines := strings.Split(text, "\n")

	for _, key := range doc.Keys {
		if key.Range.Start.Line >= len(lines) {
			continue
		}

		line := lines[key.Range.Start.Line]

		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			continue
		}

		rawValue := strings.TrimSpace(line[eqIdx+1:])
		if rawValue == "" {
			continue
		}

		var value any = rawValue

		switch {
		case strings.HasPrefix(rawValue, "["):
			value = []any{}
		case strings.HasPrefix(rawValue, "{"):
			value = map[string]any{}
		case rawValue == "true" || rawValue == "false":
			value = rawValue == "true"
		default:
			if f, err := strconv.ParseFloat(rawValue, 64); err == nil {
				if strings.Contains(rawValue, ".") {
					value = f
				} else {
					value = int64(f)
				}
			} else if looksLikeNumber(rawValue) {
				value = 0.0
			}
		}

		fullPath := strings.Join(key.Path, ".")
		values[fullPath] = value
	}

	return values
}

func looksLikeNumber(s string) bool {
	if len(s) == 0 {
		return false
	}

	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return false
	}
	firstChar := s[0]
	if (firstChar < '0' || firstChar > '9') && firstChar != '-' && firstChar != '+' {
		return false
	}

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
			continue
		}
		return false
	}

	return dotCount > 1 || strings.HasSuffix(s, ".")
}
