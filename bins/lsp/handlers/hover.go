package handlers

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/nuonco/nuon/bins/lsp/mappers"
	"github.com/nuonco/nuon/bins/lsp/models"
	tomlparser "github.com/nuonco/nuon/pkg/parser/toml"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// HoverProvider handles hover requests with injectable schema
type HoverProvider struct {
	schema          *jsonschema.Schema
	hierarchicalMap map[string]map[string]*jsonschema.Schema // Table path -> Properties mapping
	requiredMap     map[string]map[string]bool               // Table path -> Required field names
	requiredKeys    map[string]bool                          // Current context's required keys (set per hover)
}

// NewHoverProvider creates a new hover provider with a schema and builds hierarchical property map
func NewHoverProvider(schema *jsonschema.Schema) *HoverProvider {
	propMap, reqMap := mappers.BuildPropertyMap(schema)
	return &HoverProvider{
		schema:          schema,
		hierarchicalMap: propMap,
		requiredMap:     reqMap,
	}
}

// GetHoverContent returns hover information for a key in the schema
func (h *HoverProvider) GetHoverContent(text string, line, character int) *protocol.Hover {
	log.Debugf("🔍 Hover requested at line:%d char:%d", line, character)

	// Parse TOML using hybrid parser and get context
	cursorPos := tomlparser.Position{Line: line, Character: character}
	doc := tomlparser.ParseTomlWithCursor(text, cursorPos)
	tomlCtx := doc.ContextAt(cursorPos)

	log.Debugf("📍 Context detected - Table: '%s', CurrentKey: '%s', KeyPath: %v",
		tomlCtx.CurrentTable, tomlCtx.KeyOnLine, tomlCtx.KeyPath)

	if h.schema == nil || len(h.hierarchicalMap) == 0 {
		log.Warningf("⚠️  No schema available")
		return nil
	}

	if tomlCtx.KeyOnLine == "" {
		log.Debugf("📭 No hover content available at this position")
		return nil
	}

	// Lookup property in the current table level
	propertiesAtLevel, ok := h.hierarchicalMap[tomlCtx.CurrentTable]
	if !ok {
		log.Warningf("⚠️  Table '%s' not found in schema", tomlCtx.CurrentTable)
		return nil
	}

	prop, ok := propertiesAtLevel[tomlCtx.KeyOnLine]
	if !ok {
		log.Warningf("⚠️  Property '%s' not found in table '%s'", tomlCtx.KeyOnLine, tomlCtx.CurrentTable)
		return nil
	}

	if prop == nil {
		log.Warningf("⚠️  Property '%s' found but is nil", tomlCtx.KeyOnLine)
		return nil
	}

	log.Infof("✅ Found property '%s' in table '%s' (type: %s)", tomlCtx.KeyOnLine, tomlCtx.CurrentTable, prop.Type)

	// Set required keys for the current table context
	h.requiredKeys = h.requiredMap[tomlCtx.CurrentTable]

	content := h.buildHoverContent(tomlCtx.CurrentTable, tomlCtx.KeyOnLine, prop)

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: content,
		},
	}
}

// buildHoverContent formats the hover information from a property
func (h *HoverProvider) buildHoverContent(table, key string, prop *jsonschema.Schema) string {
	var content strings.Builder

	// Line 1: property name (bold) and type (italic code) on the same line
	// Bold renders in the editor's heading color, italic code renders with
	// a distinct background + slant — giving two visually different treatments.
	if prop.Type != "" {
		content.WriteString(fmt.Sprintf("**%s** &mdash; *`%s`*", key, prop.Type))
	} else {
		content.WriteString(fmt.Sprintf("**%s** &mdash; *`object`*", key))
	}

	// Append required / deprecated / default inline with the signature
	if h.isRequired(key) {
		content.WriteString(" &nbsp; `required`")
	}
	if prop.Deprecated {
		content.WriteString(" &nbsp; ~~`deprecated`~~")
	}
	if prop.Default != nil {
		content.WriteString(fmt.Sprintf(" &nbsp; default: `%v`", prop.Default))
	}
	content.WriteString("\n\n")

	// Description
	if prop.Description != "" {
		content.WriteString("---\n\n")
		content.WriteString(prop.Description)
		content.WriteString("\n\n")
	}

	// Constraints: pattern, length, numeric bounds
	if constraints := h.buildConstraints(prop); constraints != "" {
		content.WriteString(constraints)
	}

	// Enum values as inline code chips
	if len(prop.Enum) > 0 {
		content.WriteString("📋 **Allowed values:** ")
		enumStrs := make([]string, 0, len(prop.Enum))
		for _, enumVal := range prop.Enum {
			enumStrs = append(enumStrs, fmt.Sprintf("`%v`", enumVal))
		}
		content.WriteString(strings.Join(enumStrs, " · "))
		content.WriteString("\n\n")
	}

	// Child keys: for object/array types, show available sub-keys
	if childKeys := h.buildChildKeys(table, key); childKeys != "" {
		content.WriteString(childKeys)
	}

	// Examples in a TOML code block for realistic preview
	if len(prop.Examples) > 0 {
		content.WriteString("📝 **Examples:**\n\n")
		content.WriteString("```toml\n")
		for _, exampleVal := range prop.Examples {
			content.WriteString(fmt.Sprintf("%s = %v\n", key, exampleVal))
		}
		content.WriteString("```\n")
	}

	return content.String()
}

// buildConstraints returns a formatted string of schema constraints, or empty if none.
func (h *HoverProvider) buildConstraints(prop *jsonschema.Schema) string {
	var parts []string

	if prop.Pattern != "" {
		parts = append(parts, fmt.Sprintf("pattern: `%s`", prop.Pattern))
	}
	if prop.Format != "" {
		parts = append(parts, fmt.Sprintf("format: `%s`", prop.Format))
	}
	if prop.MinLength != nil {
		parts = append(parts, fmt.Sprintf("min length: `%d`", *prop.MinLength))
	}
	if prop.MaxLength != nil {
		parts = append(parts, fmt.Sprintf("max length: `%d`", *prop.MaxLength))
	}
	if prop.Minimum.String() != "" {
		parts = append(parts, fmt.Sprintf("min: `%s`", prop.Minimum))
	}
	if prop.Maximum.String() != "" {
		parts = append(parts, fmt.Sprintf("max: `%s`", prop.Maximum))
	}
	if prop.MinItems != nil {
		parts = append(parts, fmt.Sprintf("min items: `%d`", *prop.MinItems))
	}
	if prop.MaxItems != nil {
		parts = append(parts, fmt.Sprintf("max items: `%d`", *prop.MaxItems))
	}

	if len(parts) == 0 {
		return ""
	}
	return "🔒 **Constraints:** " + strings.Join(parts, " · ") + "\n\n"
}

// buildChildKeys returns a formatted list of available sub-keys for object/array types.
func (h *HoverProvider) buildChildKeys(table, key string) string {
	// Build the nested path that children would live under
	childPath := key
	if table != "" {
		childPath = table + "." + key
	}

	childProps, ok := h.hierarchicalMap[childPath]
	if !ok || len(childProps) == 0 {
		return ""
	}

	keys := make([]string, 0, len(childProps))
	for k := range childProps {
		keys = append(keys, "`"+k+"`")
	}

	return "🔑 **Sub-keys:** " + strings.Join(keys, " · ") + "\n\n"
}

// isRequired checks whether the given key is required in its current table context
func (h *HoverProvider) isRequired(key string) bool {
	if h.schema == nil {
		return false
	}
	return h.requiredKeys[key]
}

// TextDocumentHover handles hover requests
func TextDocumentHover(ctx *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	uri := params.TextDocument.URI
	pos := params.Position
	log.Infof("🔍 Hover requested at %s:%d:%d", uri, pos.Line, pos.Character)

	openDocumentsMutex.RLock()
	text, ok := openDocuments[uri]
	openDocumentsMutex.RUnlock()
	if !ok {
		log.Errorf("❌ Document not found in openDocuments: %s", uri)
		return nil, nil
	}
	log.Debugf("✅ Found document, length: %d chars", len(text))

	// Detect schema type from document
	schemaType := models.DetectSchemaType(text)
	if schemaType == "" {
		log.Warningf("⚠️  No schema type detected")
		return nil, nil
	}
	log.Debugf("✅ Detected schema type: %s", schemaType)

	// Get the schema and create hover provider
	schemaNode, err := models.LookupSchema(schemaType)
	if err != nil {
		log.Errorf("❌ Schema lookup error: %v", err)
		return nil, err
	}
	if schemaNode == nil {
		log.Warningf("⚠️  No schema node found for type '%s'", schemaType)
		return nil, nil
	}
	log.Debugf("✅ Schema node found")

	// Use HoverProvider which handles nested properties correctly
	provider := NewHoverProvider(schemaNode)
	return provider.GetHoverContent(text, int(pos.Line), int(pos.Character)), nil
}
