package handlers

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func TextDocumentDidClose(ctx *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
	log.Infof("📪 Document closed: %s", params.TextDocument.URI)
	openDocumentsMutex.Lock()
	defer openDocumentsMutex.Unlock()
	delete(openDocuments, params.TextDocument.URI)
	return nil
}
