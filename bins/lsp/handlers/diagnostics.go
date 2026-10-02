package handlers

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"

	"github.com/nuonco/nuon/bins/lsp/models"
	"github.com/nuonco/nuon/pkg/config/diagnostics"
	tomlparser "github.com/nuonco/nuon/pkg/parser/toml"
)

// PublishDiagnostics handles the full diagnostic cycle: detection, parsing, diagnosis, and publishing
func PublishDiagnostics(ctx *glsp.Context, uri protocol.DocumentUri, text string) {
	var diagnostics []protocol.Diagnostic

	// Detect schema type
	schemaType := models.DetectSchemaTypeForDocument(text, string(uri))
	if schemaType == "" {
		// Clear diagnostics if no schema detected
		ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: []protocol.Diagnostic{},
		})
		return
	}

	// Check if the schema type is valid
	if !models.IsValidSchemaType(schemaType) {
		// Find the position of the schema type on the first line
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

// DiagnoseDocument generates diagnostics for a TOML document.
func DiagnoseDocument(uri protocol.DocumentUri, doc *tomlparser.TomlDocument, rootSchema *jsonschema.Schema) []protocol.Diagnostic {
	return toProtocolDiagnostics(diagnostics.Diagnose(doc, rootSchema))
}

func extractRawValues(text string, doc *tomlparser.TomlDocument) map[string]any {
	return diagnostics.ExtractRawValues(text, doc)
}

func tomlTypeOf(value any) string {
	return diagnostics.TomlTypeOf(value)
}

func toProtocolDiagnostics(in []diagnostics.Diagnostic) []protocol.Diagnostic {
	out := make([]protocol.Diagnostic, 0, len(in))
	for _, d := range in {
		sev := protocol.DiagnosticSeverityError
		if d.Severity == diagnostics.SeverityWarning {
			sev = protocol.DiagnosticSeverityWarning
		}
		out = append(out, protocol.Diagnostic{
			Severity: ptrSeverity(sev),
			Message:  d.Message,
			Range: protocol.Range{
				Start: protocol.Position{Line: uint32(d.StartLine), Character: uint32(d.StartChar)},
				End:   protocol.Position{Line: uint32(d.EndLine), Character: uint32(d.EndChar)},
			},
			Source: ptr("Nuon LSP"),
		})
	}
	return out
}

func ptrSeverity(s protocol.DiagnosticSeverity) *protocol.DiagnosticSeverity {
	return &s
}
