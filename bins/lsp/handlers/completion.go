package handlers

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/bins/lsp/mappers"
	"github.com/nuonco/nuon/bins/lsp/models"
	tomlparser "github.com/nuonco/nuon/pkg/parser/toml"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentCompletion(ctx *glsp.Context, params *protocol.CompletionParams) (any, error) {
	uri := params.TextDocument.URI
	pos := params.Position
	log.Infof("📝 Completion requested at %s:%d:%d", uri, pos.Line, pos.Character)

	// Get the document text
	openDocumentsMutex.RLock()
	text, ok := openDocuments[uri]
	openDocumentsMutex.RUnlock()
	if !ok {
		log.Errorf("❌ Document not found in openDocuments: %s", uri)
		return nil, fmt.Errorf("document not found: %s", uri)
	}
	log.Debugf("✅ Found document, length: %d chars", len(text))

	// Check if we're on line 0 (first line) after a # - provide schema type completions
	if pos.Line == 0 {
		lines := strings.Split(text, "\n")
		if len(lines) > 0 {
			firstLine := lines[0]
			beforeCursor := firstLine[:min(int(pos.Character), len(firstLine))]
			if strings.HasPrefix(strings.TrimSpace(beforeCursor), "#") {
				log.Infof("🏷️  Building schema type completions for first line")
				return buildSchemaTypeCompletions(), nil
			}
		}
	}

	// Parse TOML using hybrid parser
	cursorPos := tomlparser.Position{Line: int(pos.Line), Character: int(pos.Character)}
	doc := tomlparser.ParseTomlWithCursor(text, cursorPos)
	log.Debugf("✅ TOML parsed, %d tables, %d keys found", len(doc.Tables), len(doc.Keys))

	// Get context at cursor position
	tomlCtx := doc.ContextAt(cursorPos)
	log.Infof("📍 Context detected - Table: '%s', CurrentKey: '%s', KeyPath: %v",
		tomlCtx.CurrentTable, tomlCtx.KeyOnLine, tomlCtx.KeyPath)

	// Detect schema type from document
	schemaType := models.DetectSchemaType(text)
	if schemaType == "" {
		log.Warningf("⚠️  No schema type detected, returning no completions")
		return nil, nil
	}
	log.Debugf("✅ Detected schema type: %s", schemaType)

	// Get schema for detected type
	schema, err := models.LookupSchema(schemaType)
	if err != nil {
		log.Errorf("❌ Schema lookup error: %v", err)
		return nil, err
	}
	if schema == nil {
		log.Warningf("⚠️  No schema found for type '%s'", schemaType)
		return nil, nil
	}

	// Build hierarchical property map from schema
	hierarchicalMap := mappers.BuildPropertyMap(schema)
	log.Debugf("✅ Built hierarchical property map with %d table levels", len(hierarchicalMap))

	// Build completions
	var items []protocol.CompletionItem

	// Determine if cursor is on a key or value
	lines := strings.Split(text, "\n")
	currentLine := ""
	if int(pos.Line) < len(lines) {
		currentLine = lines[pos.Line]
	}
	beforeCursor := currentLine[:min(int(pos.Character), len(currentLine))]
	isValue := strings.Contains(beforeCursor, "=")

	if !isValue {
		log.Infof("🔑 Building KEY completions for table '%s'", tomlCtx.CurrentTable)
		// Key context → suggest properties from current table level
		propertiesAtLevel, ok := hierarchicalMap[tomlCtx.CurrentTable]
		if !ok || len(propertiesAtLevel) == 0 {
			log.Warningf("⚠️  No properties found for table '%s'", tomlCtx.CurrentTable)
		} else {
			count := 0
			for key, prop := range propertiesAtLevel {
				docStr := ""
				if prop != nil {
					docStr = prop.Description
				}
				detail := ""
				if prop != nil && prop.Type != "" {
					detail = prop.Type
				}
				kind := protocol.CompletionItemKindProperty
				items = append(items, protocol.CompletionItem{
					Label:         key,
					Kind:          &kind,
					Detail:        &detail,
					Documentation: &protocol.MarkupContent{Kind: protocol.MarkupKindMarkdown, Value: docStr},
					InsertText:    ptr(fmt.Sprintf("%s = ", key)),
				})
				count++
			}
			log.Infof("✅ Generated %d key completions", count)
		}
	} else if tomlCtx.KeyOnLine != "" {
		log.Infof("💡 Building VALUE completions for key '%s' in table '%s'", tomlCtx.KeyOnLine, tomlCtx.CurrentTable)
		// Value context → suggest enum/examples/defaults for that key
		// Look up the property in the current table level
		propertiesAtLevel, ok := hierarchicalMap[tomlCtx.CurrentTable]
		if !ok {
			log.Warningf("⚠️  Table '%s' not found in schema", tomlCtx.CurrentTable)
		} else {
			prop, ok := propertiesAtLevel[tomlCtx.KeyOnLine]
			if ok && prop != nil {
				log.Debugf("✅ Found property schema for '%s' (type: %s)", tomlCtx.KeyOnLine, prop.Type)
				// Handle boolean type suggestions
				if prop.Type == "boolean" {
					items = append(items,
						protocol.CompletionItem{Label: "true"},
						protocol.CompletionItem{Label: "false"},
					)
					log.Infof("✅ Generated 2 boolean value completions")
				} else if len(prop.Enum) > 0 {
					// Handle enum values
					isStringType := prop.Type == "string"
					for _, enumVal := range prop.Enum {
						value := fmt.Sprintf("%v", enumVal)
						if isStringType {
							value = fmt.Sprintf("\"%s\"", value)
						}
						items = append(items, protocol.CompletionItem{
							Label:      fmt.Sprintf("%v", enumVal),
							InsertText: &value,
						})
					}
					log.Infof("✅ Generated %d enum value completions", len(prop.Enum))
				} else if len(prop.Examples) > 0 {
					// Handle examples for other types
					isStringType := prop.Type == "string"
					for _, exampleVal := range prop.Examples {
						value := fmt.Sprintf("%v", exampleVal)
						if isStringType {
							value = fmt.Sprintf("\"%s\"", value)
						}
						items = append(items, protocol.CompletionItem{
							Label:      fmt.Sprintf("%v", exampleVal),
							InsertText: &value,
						})
					}
					log.Infof("✅ Generated %d example value completions", len(prop.Examples))
				} else {
					log.Warningf("⚠️  No completions available for property '%s' (no enum/examples)", tomlCtx.KeyOnLine)
				}
			} else {
				log.Warningf("⚠️  Property '%s' not found in schema", tomlCtx.KeyOnLine)
			}
		}
	}

	log.Infof("🎯 Returning %d total completion items", len(items))
	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// small helper to get pointer of string
func ptr(s string) *string { return &s }

// buildSchemaTypeCompletions returns completion items for all valid schema types
func buildSchemaTypeCompletions() *protocol.CompletionList {
	schemaTypes := models.GetValidSchemaTypes()
	items := make([]protocol.CompletionItem, 0, len(schemaTypes))

	for _, schemaType := range schemaTypes {
		kind := protocol.CompletionItemKindValue
		detail := "Schema type"
		items = append(items, protocol.CompletionItem{
			Label:      schemaType,
			Kind:       &kind,
			Detail:     &detail,
			InsertText: ptr(schemaType),
		})
	}

	log.Infof("✅ Generated %d schema type completions", len(items))
	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}
