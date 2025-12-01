package handlers

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/powertoolsdev/mono/bins/lsp/mappers"
	"github.com/powertoolsdev/mono/bins/lsp/models"
	tomlparser "github.com/powertoolsdev/mono/pkg/parser/toml"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// HoverProvider handles hover requests with injectable schema
type HoverProvider struct {
	schema          *jsonschema.Schema
	hierarchicalMap map[string]map[string]*jsonschema.Schema // Table path -> Properties mapping
}

// NewHoverProvider creates a new hover provider with a schema and builds hierarchical property map
func NewHoverProvider(schema *jsonschema.Schema) *HoverProvider {
	return &HoverProvider{
		schema:          schema,
		hierarchicalMap: mappers.BuildPropertyMap(schema),
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
	content := h.buildHoverContent(tomlCtx.KeyOnLine, prop)

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: content,
		},
	}
}

// buildHoverContent formats the hover information from a property
func (h *HoverProvider) buildHoverContent(key string, prop *jsonschema.Schema) string {
	var content strings.Builder
	if prop.Type != "" {
		content.WriteString(fmt.Sprintf("**%s** (`%s`)\n\n", key, prop.Type))
	} else {
		content.WriteString(fmt.Sprintf("**%s**\n\n", key))
	}

	if prop.Description != "" {
		content.WriteString(fmt.Sprintf("%s\n\n", prop.Description))
	}

	if len(prop.Enum) > 0 {
		content.WriteString("**Enum values:**\n")
		for _, enumVal := range prop.Enum {
			content.WriteString(fmt.Sprintf("- `%v`\n", enumVal))
		}
		content.WriteString("\n")
	}

	if len(prop.Examples) > 0 {
		content.WriteString("**Examples:**\n")
		for _, exampleVal := range prop.Examples {
			content.WriteString(fmt.Sprintf("- `%v`\n", exampleVal))
		}
	}

	return content.String()
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
