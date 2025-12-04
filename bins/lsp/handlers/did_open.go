package handlers

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	text := params.TextDocument.Text
	log.Infof("📂 Document opened: %s (length: %d chars)", uri, len(text))

	openDocumentsMutex.Lock()
	openDocuments[uri] = text
	openDocumentsMutex.Unlock()

	// Trigger diagnostics
	PublishDiagnostics(ctx, uri, text)

	return nil
}
